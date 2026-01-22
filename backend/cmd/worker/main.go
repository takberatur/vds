package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
	"github.com/user/video-downloader-backend/pkg/logger"
)

func main() {
	logger.Init()
	log.Info().Msg("Starting worker...")

	cfg := config.LoadConfig()

	db, err := infrastructure.NewPostgresClient(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Pool.Close()

	downloadRepo := repository.NewDownloadRepository(db.Pool)
	downloader := infrastructure.NewFallbackDownloader()

	storageClient, err := infrastructure.NewStorageClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create storage client")
	}

	// Ensure bucket exists
	if err := storageClient.CreateBucket(context.Background(), cfg.MinioBucket); err != nil {
		log.Fatal().Err(err).Msg("failed to create minio bucket")
	}

	// Start Cleanup Cron Job
	go startCleanupCron(context.Background(), downloadRepo, storageClient, cfg.MinioBucket)

	server := infrastructure.NewTaskServer(cfg.RedisAddr, cfg.RedisPassword)

	redisClient, err := infrastructure.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis for events")
	}
	defer redisClient.Close()

	mux := asynq.NewServeMux()

	mux.HandleFunc(infrastructure.TypeVideoDownload, func(ctx context.Context, t *asynq.Task) error {
		var task model.DownloadTask
		if err := json.Unmarshal(t.Payload(), &task); err != nil {
			return err
		}

		if err := handleVideoDownloadTask(ctx, downloadRepo, redisClient, downloader, storageClient, cfg.MinioBucket, &task); err != nil {
			return err
		}

		return nil
	})

	if err := server.Run(mux); err != nil {
		log.Fatal().Err(err).Msg("asynq server stopped with error")
	}
}

func startCleanupCron(ctx context.Context, downloadRepo repository.DownloadRepository, storageClient infrastructure.StorageClient, bucketName string) {
	// Run every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	log.Info().Msg("Cleanup cron job initialized (interval: 10m, retention: 30m)")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Info().Msg("Starting cleanup cron job execution")
			// Cleanup tasks older than 30 minutes
			cutoff := time.Now().Add(-30 * time.Minute)

			// Fetch in batches
			for {
				tasks, err := downloadRepo.FindOldAndCompleted(ctx, cutoff, 100)
				if err != nil {
					log.Error().Err(err).Msg("Failed to find old tasks for cleanup")
					break
				}
				if len(tasks) == 0 {
					break
				}

				var idsToDelete []uuid.UUID
				for _, task := range tasks {
					// Delete files from MinIO
					// Folder structure: platform_type/task_id/
					prefix := fmt.Sprintf("%s/%s/", task.PlatformType, task.ID.String())
					if err := storageClient.DeleteFolder(ctx, bucketName, prefix); err != nil {
						log.Error().Err(err).Str("task_id", task.ID.String()).Msg("Failed to delete folder from MinIO")
						// We continue to delete from DB to avoid orphan records, 
						// or we could skip adding to idsToDelete. 
						// Let's delete from DB to keep it clean as requested.
					}
					idsToDelete = append(idsToDelete, task.ID)
				}

				if len(idsToDelete) > 0 {
					if err := downloadRepo.BulkDelete(ctx, idsToDelete); err != nil {
						log.Error().Err(err).Msg("Failed to bulk delete tasks from DB")
					} else {
						log.Info().Int("count", len(idsToDelete)).Msg("Deleted old tasks and files")
					}
				}
			}
		}
	}
}

func publishDownloadEvent(ctx context.Context, redisClient infrastructure.RedisClient, event *model.DownloadEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return redisClient.Publish(ctx, infrastructure.DownloadEventChannel, data).Err()
}

func handleVideoDownloadTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, downloader infrastructure.DownloaderClient, storageClient infrastructure.StorageClient, bucketName string, task *model.DownloadTask) error {
	task.Status = "processing"
	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to processing")
	}

	if err := publishStartEvent(ctx, redisClient, task); err != nil {
		log.Error().Err(err).Msg("failed to publish start event")
	}

	if err := processDownloadTask(ctx, downloadRepo, redisClient, downloader, storageClient, bucketName, task); err != nil {
		failErr := markTaskFailed(ctx, downloadRepo, redisClient, task, err)
		if failErr != nil {
			log.Error().Err(failErr).Str("task_id", task.ID.String()).Msg("failed to mark task as failed")
		}
		return err
	}

	return nil
}

func processDownloadTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, downloader infrastructure.DownloaderClient, storageClient infrastructure.StorageClient, bucketName string, task *model.DownloadTask) error {
	if err := publishProgressEvent(ctx, redisClient, task, 10); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", 10).Msg("failed to publish progress event (start)")
	}

	// 1. Get Video Info
	info, err := downloader.GetVideoInfo(ctx, task.OriginalURL)
	if err != nil {
		return err
	}

	// 2. Update Task Metadata
	if info != nil {
		if info.Title != "" {
			t := info.Title
			task.Title = &t
		}
		if info.Thumbnail != "" {
			t := info.Thumbnail
			task.ThumbnailURL = &t
		}
		if info.Duration > 0 {
			d := int(info.Duration)
			task.Duration = &d
		}
	}

	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task metadata")
	}

	if err := publishProgressEvent(ctx, redisClient, task, 30); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", 30).Msg("failed to publish progress event (metadata)")
	}

	// 3. Select Formats to Download
	// For now, let's pick the best video format and maybe some common ones
	// Or just download the best one first as requested, and then handle others if needed.
	// The user wants "pilihan semua format jika ada".
	// Let's filter formats to get unique resolutions (e.g. 1080p, 720p, 480p, 360p)
	selectedFormats := pickFormatsToDownload(info.Formats)

	// If no formats found but we have a download URL, use it
	if len(selectedFormats) == 0 && info.DownloadURL != "" {
		selectedFormats = append(selectedFormats, infrastructure.FormatInfo{
			URL: info.DownloadURL,
			Ext: "mp4",
		})
	}

	for i, fmtInfo := range selectedFormats {
		progress := 30 + int(float64(i)/float64(len(selectedFormats))*50)
		if err := publishProgressEvent(ctx, redisClient, task, progress); err != nil {
			log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", progress).Msg("failed to publish progress event (downloading)")
		}

		// 4. Download each format
		// Use resolution in temp filename to avoid invalid chars from selector (e.g. "/")
		tempPattern := fmt.Sprintf("vid-%dp-*.%s", fmtInfo.Height, fmtInfo.Ext)
		if fmtInfo.Height == 0 {
			tempPattern = fmt.Sprintf("vid-best-*.%s", fmtInfo.Ext)
		}

		tempFile, err := os.CreateTemp("", tempPattern)
		if err != nil {
			log.Error().Err(err).Msg("failed to create temp file")
			continue
		}
		tempPath := tempFile.Name()
		tempFile.Close()
		defer os.Remove(tempPath)

		err = downloader.DownloadToPath(ctx, task.OriginalURL, fmtInfo.FormatID, tempPath)
		if err != nil {
			log.Error().Err(err).Str("format", fmtInfo.FormatID).Msg("failed to download format")
			continue
		}

		// 5. Upload to MinIO
		f, err := os.Open(tempPath)
		if err != nil {
			log.Error().Err(err).Msg("failed to open temp file for upload")
			continue
		}

		fi, _ := f.Stat()
		resolution := ""
		if fmtInfo.Height > 0 {
			resolution = fmt.Sprintf("%dp", fmtInfo.Height)
		}

		// Use resolution for filename to avoid special chars from format selector
		safeName := resolution
		if safeName == "" {
			safeName = "best"
		}
		objectName := fmt.Sprintf("%s/%s/%s.%s", task.PlatformType, task.ID.String(), safeName, fmtInfo.Ext)

		contentType := "video/mp4"
		if fmtInfo.Ext == "webm" {
			contentType = "video/webm"
		}

		minioURL, err := storageClient.UploadFile(ctx, bucketName, objectName, f, fi.Size(), contentType)
		f.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to upload to minio")
			continue
		}

		// 6. Save to download_files
		ext := fmtInfo.Ext
		size := fi.Size()
		// For DB, we can store the resolution as format_id if actual ID is complex selector
		fID := safeName
		res := resolution

		downloadFile := &model.DownloadFile{
			DownloadID: task.ID,
			URL:        minioURL,
			FormatID:   &fID,
			Resolution: &res,
			Extension:  &ext,
			FileSize:   &size,
		}
		if err := downloadRepo.AddFile(ctx, downloadFile); err != nil {
			log.Error().Err(err).Msg("failed to add download file record")
		}

		// Set primary file path to the first (usually best) format
		if i == 0 {
			task.FilePath = &minioURL
			task.FileSize = &size
			task.Format = &fID
		}
	}

	// 7. Update Task Status
	task.Status = "completed"
	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to completed")
	}

	// 8. Publish Completion
	if err := publishCompletionEvent(ctx, redisClient, task); err != nil {
		log.Error().Err(err).Msg("failed to publish complete event")
	}

	return nil
}

func pickFormatsToDownload(formats []infrastructure.FormatInfo) []infrastructure.FormatInfo {
	if len(formats) == 0 {
		return nil
	}

	// Identify available heights from video streams
	// We want to capture the best quality for each resolution bucket
	heights := make(map[int]bool)
	for _, f := range formats {
		// Filter out audio-only (vcodec=none) and very low quality
		if f.Height >= 360 && f.Vcodec != "none" {
			heights[f.Height] = true
		}
	}

	var hList []int
	for h := range heights {
		hList = append(hList, h)
	}

	// Sort descending
	sort.Sort(sort.Reverse(sort.IntSlice(hList)))

	// Limit to top 4 resolutions
	if len(hList) > 4 {
		hList = hList[:4]
	}

	var result []infrastructure.FormatInfo
	for _, h := range hList {
		// Construct selector for "best video at height H + best audio"
		// This ensures we get combined video/audio file
		selector := fmt.Sprintf("bestvideo[height=%d]+bestaudio/best[height=%d]", h, h)

		result = append(result, infrastructure.FormatInfo{
			FormatID: selector,
			Height:   h,
			Ext:      "mp4", // We'll let yt-dlp merge to mp4 (requires ffmpeg)
		})
	}

	return result
}

func intPtr(v int) *int {
	return &v
}

func publishStartEvent(ctx context.Context, redisClient infrastructure.RedisClient, task *model.DownloadTask) error {
	event := &model.DownloadEvent{
		Type:      "download.processing",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "processing",
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, event)
}

func publishProgressEvent(ctx context.Context, redisClient infrastructure.RedisClient, task *model.DownloadTask, progress int) error {
	event := &model.DownloadEvent{
		Type:      "download.processing",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "processing",
		Progress:  &progress,
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, event)
}

func publishCompletionEvent(ctx context.Context, redisClient infrastructure.RedisClient, task *model.DownloadTask) error {
	var payload *model.DownloadPayload

	// Prepare payload file path
	var payloadFilePath *string
	if task.FilePath != nil {
		payloadFilePath = task.FilePath
	}

	// Prepare formats for payload
	var payloadFormats []model.DownloadFormat
	for _, f := range task.DownloadFiles {
		format := model.DownloadFormat{
			URL: f.URL,
		}
		if f.FormatID != nil {
			format.FormatID = *f.FormatID
		}
		if f.Extension != nil {
			format.Ext = *f.Extension
		}
		if f.FileSize != nil {
			format.Filesize = *f.FileSize
		}
		if f.Resolution != nil {
			// Extract height from resolution string like "1080p"
			var h int
			fmt.Sscanf(*f.Resolution, "%dp", &h)
			format.Height = h
		}
		payloadFormats = append(payloadFormats, format)
	}

	payload = &model.DownloadPayload{
		FilePath: payloadFilePath,
		Formats:  payloadFormats,
	}

	log.Info().
		Str("task_id", task.ID.String()).
		Int("formats_count", len(payloadFormats)).
		Msg("Publishing completion event with MinIO URLs")

	event := &model.DownloadEvent{
		Type:      "download.completed",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "completed",
		Progress:  intPtr(100),
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, event)
}

func markTaskFailed(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, task *model.DownloadTask, err error) error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}

	task.Status = "failed"
	task.ErrorMessage = &msg

	if updateErr := downloadRepo.Update(ctx, task); updateErr != nil {
		log.Error().Err(updateErr).Str("task_id", task.ID.String()).Msg("failed to update task to failed")
	}

	event := &model.DownloadEvent{
		Type:      "download.failed",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "failed",
		Error:     msg,
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, event)
}
