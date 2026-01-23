package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
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
		if info.Duration != nil && *info.Duration > 0 {
			d := int(*info.Duration)
			task.Duration = &d
		}
	}

	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task metadata")
	}

	if err := publishProgressEvent(ctx, redisClient, task, 30); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", 30).Msg("failed to publish progress event (metadata)")
	}

	// Special handling for platforms with direct URLs (YouTube, Facebook, Twitter/X)
	// We skip the download-upload loop and just save the direct URLs
	isTwitter := strings.ToLower(task.PlatformType) == "twitter" ||
		strings.ToLower(task.PlatformType) == "x" ||
		strings.Contains(strings.ToLower(task.OriginalURL), "twitter.com") ||
		strings.Contains(strings.ToLower(task.OriginalURL), "x.com") ||
		strings.Contains(strings.ToLower(task.OriginalURL), "twimg.com")

	isInstagram := strings.ToLower(task.PlatformType) == "instagram" ||
		strings.Contains(strings.ToLower(task.OriginalURL), "instagram.com")

	isTiktok := strings.ToLower(task.PlatformType) == "tiktok" ||
		strings.Contains(strings.ToLower(task.OriginalURL), "tiktok.com")

	if strings.ToLower(task.PlatformType) == "youtube" ||
		strings.ToLower(task.PlatformType) == "facebook" ||
		isTwitter ||
		isInstagram ||
		isTiktok {
		log.Info().Str("platform", task.PlatformType).Msg("Processing as direct download (no-upload)")
		return processDirectLinkTask(ctx, downloadRepo, redisClient, task, info)
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
		if fmtInfo.Height == nil || *fmtInfo.Height == 0 {
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
		if fmtInfo.Height != nil && *fmtInfo.Height > 0 {
			resolution = fmt.Sprintf("%dp", *fmtInfo.Height)
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

func processDirectLinkTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, task *model.DownloadTask, info *infrastructure.VideoInfo) error {
	// 0. Ensure platform type is correct before we start
	// This helps with Twitter detection if it was missed earlier
	lowerURL := strings.ToLower(task.OriginalURL)
	if task.PlatformType == "" {
		if strings.Contains(lowerURL, "twitter.com") || strings.Contains(lowerURL, "x.com") || strings.Contains(lowerURL, "twimg.com") {
			task.PlatformType = "twitter"
		} else if strings.Contains(lowerURL, "youtube.com") || strings.Contains(lowerURL, "youtu.be") {
			task.PlatformType = "youtube"
		} else if strings.Contains(lowerURL, "facebook.com") || strings.Contains(lowerURL, "fb.watch") {
			task.PlatformType = "facebook"
		} else if strings.Contains(lowerURL, "instagram.com") {
			task.PlatformType = "instagram"
		} else if strings.Contains(lowerURL, "tiktok.com") {
			task.PlatformType = "tiktok"
		}
	}

	/*
		isTwitter := strings.EqualFold(task.PlatformType, "twitter") ||
			strings.EqualFold(task.PlatformType, "x") ||
			strings.Contains(lowerURL, "twitter.com") ||
			strings.Contains(lowerURL, "x.com") ||
			strings.Contains(lowerURL, "twimg.com")
	*/

	/*
		isInstagram := strings.EqualFold(task.PlatformType, "instagram") ||
			strings.Contains(lowerURL, "instagram.com")
	*/

	isTiktok := strings.EqualFold(task.PlatformType, "tiktok") ||
		strings.Contains(lowerURL, "tiktok.com")

	// Helper function to clean URLs
	cleanURL := func(u string) string {
		// Always clean surrounding whitespace and backticks which are invalid in URLs
		// This fixes issues where URLs might be wrapped in backticks from logging or scraping artifacts
		u = strings.TrimSpace(u)
		u = strings.Trim(u, "`")
		u = strings.ReplaceAll(u, "`", "")
		u = strings.Trim(u, "'")
		u = strings.Trim(u, "\"")

		// Note: We do NOT remove query params (e.g. ?tag=21) as they are required for
		// Twitter, Instagram, and TikTok access.
		return u
	}

	// Sanitize task URLs immediately
	if task.ThumbnailURL != nil {
		cleaned := cleanURL(*task.ThumbnailURL)
		task.ThumbnailURL = &cleaned
	}
	if info != nil {
		if info.Thumbnail != "" {
			info.Thumbnail = cleanURL(info.Thumbnail)
		}
		if info.DownloadURL != "" {
			info.DownloadURL = cleanURL(info.DownloadURL)
		}
	}

	// 1. Gather all formats (or filtered)
	// We want to save ALL available formats that have a valid URL
	var formatsToSave []infrastructure.FormatInfo
	if len(info.Formats) > 0 {
		// Filter for YouTube/Facebook/Twitter/Instagram/TikTok:
		// Keep formats, but maybe add metadata to indicate video-only/audio-only if needed in future
		var validFormats []infrastructure.FormatInfo
		for _, f := range info.Formats {
			// Restore Filter: Only keep formats with both Video and Audio
			// This prevents "video only" or "audio only" links which are useless as direct downloads for average users

			// Relaxed check: empty string often means "present but unknown" in some contexts,
			// while "none" explicitly means missing.
			hasVideo := f.Vcodec != "none"
			hasAudio := f.Acodec != "none"

			// If specific codecs are empty, we give benefit of doubt for HTTP formats (often combined)
			// checking ext might help (mp4 usually has both unless specified)
			if f.Vcodec == "" && f.Acodec == "" && f.Ext == "mp4" {
				hasVideo = true
				hasAudio = true
			}

			// Special case for TikTok: yt-dlp might return formats with weird codec strings or "none"
			// but if it has a valid http URL and it's not m3u8, it's likely playable.
			if isTiktok && (strings.HasPrefix(f.URL, "http") && !strings.Contains(f.URL, ".m3u8")) {
				// Assume playable if it's a direct http link for tiktok
				hasVideo = true
				hasAudio = true
			}

			if hasVideo && hasAudio {
				// Clean URL before saving
				f.URL = cleanURL(f.URL)
				validFormats = append(validFormats, f)
			}
		}

		// Use filtered formats if any found, otherwise fallback to all
		if len(validFormats) > 0 {
			formatsToSave = validFormats
		} else {
			formatsToSave = info.Formats
		}
	} else if info.DownloadURL != "" {
		formatsToSave = append(formatsToSave, infrastructure.FormatInfo{
			URL: cleanURL(info.DownloadURL),
			Ext: "mp4", // Default guess
		})
	}

	// 2. Save to download_files table
	// We also need to find the "best" format to set as the main file for the task
	var bestFile *model.DownloadFile

	for i, f := range formatsToSave {
		if f.URL == "" {
			continue
		}

		// Construct DB model
		// Use safe defaults for nil pointers
		var resolution string
		if f.Height != nil && *f.Height > 0 {
			resolution = fmt.Sprintf("%dp", *f.Height)
		} else if f.Width != nil && *f.Width > 0 {
			resolution = fmt.Sprintf("%dw", *f.Width) // fallback
		}

		// Format ID from yt-dlp (e.g. "137", "22", "sb3")
		fmtID := f.FormatID
		if fmtID == "" {
			fmtID = "unknown"
		}

		ext := f.Ext
		if ext == "" {
			ext = "unknown"
		}

		var size int64
		if f.Filesize != nil {
			size = *f.Filesize
		}

		downloadFile := &model.DownloadFile{
			DownloadID: task.ID,
			URL:        f.URL,
			FormatID:   &fmtID,
			Resolution: &resolution,
			Extension:  &ext,
			FileSize:   &size,
		}

		// Save to DB
		if err := downloadRepo.AddFile(ctx, downloadFile); err != nil {
			log.Error().Err(err).Str("format_id", fmtID).Msg("failed to add direct download file record")
			// Continue to try other formats
		}

		// Append to task.DownloadFiles so publishCompletionEvent includes them
		task.DownloadFiles = append(task.DownloadFiles, *downloadFile)

		// Determine if this is the "best" file to represent the task
		// Heuristic: Highest resolution, or largest file size
		// Simple logic: if this is the first one, or better than current best
		if bestFile == nil {
			bestFile = downloadFile
		} else {
			// Compare resolution if available
			if f.Height != nil && bestFile.Resolution != nil {
				var currentH int
				fmt.Sscanf(*bestFile.Resolution, "%dp", &currentH)
				if *f.Height > currentH {
					bestFile = downloadFile
				}
			}
		}

		// Progress simulation?
		if i%5 == 0 {
			progress := 30 + int(float64(i)/float64(len(formatsToSave))*60)
			if progress > 99 {
				progress = 99
			}
			publishProgressEvent(ctx, redisClient, task, progress)
		}
	}

	// 3. Update Task Status & Main File Info
	if bestFile != nil {
		task.FilePath = &bestFile.URL
		task.FileSize = bestFile.FileSize
		task.Format = bestFile.FormatID
	} else if info.DownloadURL != "" {
		// Fallback if no formats loop worked but top level has URL
		u := info.DownloadURL
		task.FilePath = &u
	}

	task.Status = "completed"
	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to completed")
		return err
	}

	// 4. Publish Completion
	if err := publishCompletionEvent(ctx, redisClient, task); err != nil {
		log.Error().Err(err).Msg("failed to publish complete event")
	}

	log.Info().Str("task_id", task.ID.String()).Int("files_count", len(task.DownloadFiles)).Msg("Direct download processing completed")
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
		if f.Height != nil && *f.Height >= 360 && f.Vcodec != "none" {
			heights[*f.Height] = true
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
			Height:   intPtr(h),
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
			URL:      f.URL,
			FormatID: getValueOrEmpty(f.FormatID),
			Ext:      getValueOrEmpty(f.Extension),
		}

		if f.FileSize != nil {
			format.Filesize = f.FileSize
		}

		// Try to parse height from resolution if possible, or leave it nil
		if f.Resolution != nil {
			var h int
			if _, err := fmt.Sscanf(*f.Resolution, "%dp", &h); err == nil {
				format.Height = &h
			}
		}

		payloadFormats = append(payloadFormats, format)
	}

	// Ensure platform type is set
	platformType := task.PlatformType
	if platformType == "" {
		// Fallback detection from OriginalURL
		lowerURL := strings.ToLower(task.OriginalURL)
		if strings.Contains(lowerURL, "twitter.com") || strings.Contains(lowerURL, "x.com") {
			platformType = "twitter"
		} else if strings.Contains(lowerURL, "youtube.com") || strings.Contains(lowerURL, "youtu.be") {
			platformType = "youtube"
		} else if strings.Contains(lowerURL, "facebook.com") || strings.Contains(lowerURL, "fb.watch") {
			platformType = "facebook"
		} else if strings.Contains(lowerURL, "instagram.com") {
			platformType = "instagram"
		} else if strings.Contains(lowerURL, "tiktok.com") {
			platformType = "tiktok"
		}
	}

	payload = &model.DownloadPayload{
		ID:           task.ID,
		Status:       task.Status,
		Progress:     100,
		Title:        getValueOrEmpty(task.Title),
		ThumbnailURL: getValueOrEmpty(task.ThumbnailURL),
		Type:         platformType,
		CreatedAt:    task.CreatedAt,
		FilePath:     payloadFilePath,
		Formats:      payloadFormats,
	}

	event := &model.DownloadEvent{
		Type:      "download.completed",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "completed",
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, event)
}

func getValueOrEmpty(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func markTaskFailed(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, task *model.DownloadTask, err error) error {
	task.Status = "failed"
	errMsg := err.Error()
	task.ErrorMessage = &errMsg

	if updateErr := downloadRepo.Update(ctx, task); updateErr != nil {
		return updateErr
	}

	event := &model.DownloadEvent{
		Type:      "download.failed",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "failed",
		Error:     errMsg,
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, event)
}
