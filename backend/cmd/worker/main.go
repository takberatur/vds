package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

		if err := handleVideoDownloadTask(ctx, downloadRepo, redisClient, downloader, &task); err != nil {
			return err
		}

		return nil
	})

	if err := server.Run(mux); err != nil {
		log.Fatal().Err(err).Msg("asynq server stopped with error")
	}
}

func publishDownloadEvent(ctx context.Context, redisClient infrastructure.RedisClient, event *model.DownloadEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return redisClient.Publish(ctx, infrastructure.DownloadEventChannel, data).Err()
}

func handleVideoDownloadTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, downloader infrastructure.DownloaderClient, task *model.DownloadTask) error {
	task.Status = "processing"
	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to processing")
	}

	if err := publishStartEvent(ctx, redisClient, task); err != nil {
		log.Error().Err(err).Msg("failed to publish start event")
	}

	if err := processDownloadTask(ctx, downloadRepo, redisClient, downloader, task); err != nil {
		failErr := markTaskFailed(ctx, downloadRepo, redisClient, task, err)
		if failErr != nil {
			log.Error().Err(failErr).Str("task_id", task.ID.String()).Msg("failed to mark task as failed")
		}
		return err
	}

	return nil
}

func processDownloadTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, downloader infrastructure.DownloaderClient, task *model.DownloadTask) error {
	if err := publishProgressEvent(ctx, redisClient, task, 10); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", 10).Msg("failed to publish progress event (start)")
	}

	// Optimization: If the task already has a file path (download URL) and title,
	// we assume the service already fetched the info successfully.
	// We don't need to re-fetch it in the worker, which might fail or be redundant.
	// However, if FilePath equals OriginalURL, it might mean we just have the page URL, so we should re-fetch
	// unless it's a lux:// URL which is definitely processed.
	alreadyHasInfo := task.FilePath != nil && *task.FilePath != "" &&
		task.Title != nil && *task.Title != "" &&
		(*task.FilePath != task.OriginalURL || strings.HasPrefix(*task.FilePath, "lux://"))

	var info *infrastructure.VideoInfo
	var err error

	if !alreadyHasInfo {
		info, err = downloader.GetVideoInfo(ctx, task.OriginalURL)
		if err != nil {
			return err
		}
	} else {
		log.Info().Str("task_id", task.ID.String()).Msg("Task already has video info, skipping re-fetch")
	}

	if info != nil {
		if info.Duration > 0 {
			dur := int(info.Duration)
			task.Duration = &dur
		}
		if info.Filesize > 0 {
			size := info.Filesize
			task.FileSize = &size
		}
		if info.Title != "" {
			title := info.Title
			task.Title = &title
		}
		if info.Thumbnail != "" {
			thumb := info.Thumbnail
			task.ThumbnailURL = &thumb
		}
		if info.DownloadURL != "" {
			path := info.DownloadURL
			task.FilePath = &path
		}
		if len(info.Formats) > 0 {
			formats := make([]model.DownloadFormat, 0, len(info.Formats))
			for _, f := range info.Formats {
				if f.URL == "" {
					continue
				}
				formats = append(formats, model.DownloadFormat{
					URL:      f.URL,
					Filesize: f.Filesize,
					FormatID: f.FormatID,
					Acodec:   f.Acodec,
					Vcodec:   f.Vcodec,
					Ext:      f.Ext,
					Height:   f.Height,
					Width:    f.Width,
					Tbr:      f.Tbr,
				})
			}
			task.Formats = formats
		}
	}

	// Ensure formats are populated if we have a file path but no formats
	// This handles cases where we skipped fetching info (alreadyHasInfo) OR where info didn't return formats (e.g. direct link)
	if len(task.Formats) == 0 && task.FilePath != nil {
		task.Formats = []model.DownloadFormat{
			{
				FormatID: "download", // Use "download" so handler ignores it and uses default logic
				URL:      *task.FilePath,
				Ext:      "mp4", // Default to mp4 as safe bet for video players
			},
		}
	}

	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task metadata")
	}

	if err := publishProgressEvent(ctx, redisClient, task, 60); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", 60).Msg("failed to publish progress event (metadata)")
	}

	task.Status = "completed"
	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to completed")
	}

	if err := publishCompletionEvent(ctx, redisClient, task); err != nil {
		log.Error().Err(err).Msg("failed to publish complete event")
	}

	return nil
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

	// Prepare payload formats with proxy URLs
	var payloadFormats []model.DownloadFormat
	if len(task.Formats) > 0 {
		payloadFormats = make([]model.DownloadFormat, len(task.Formats))
		copy(payloadFormats, task.Formats)

		for i := range payloadFormats {
			// Convert all URLs to backend proxy URLs to handle Lux scheme, CORS, and headers
			payloadFormats[i].URL = fmt.Sprintf("/api/v1/public-proxy/downloads/file?task_id=%s&format_id=%s", task.ID.String(), payloadFormats[i].FormatID)
		}
	}

	// Prepare payload file path with proxy URL
	var payloadFilePath *string
	if task.FilePath != nil {
		// Always use proxy URL for consistency and to hide internal schemes/direct links
		// EXCEPT if the task.FilePath itself is ALREADY a proxy URL (recursion guard)
		if strings.Contains(*task.FilePath, "/api/v1/public-proxy/") {
			log.Warn().Str("task_id", task.ID.String()).Msg("Task FilePath is already a proxy URL, using it as is")
			payloadFilePath = task.FilePath
		} else {
			proxyURL := fmt.Sprintf("/api/v1/public-proxy/downloads/file?task_id=%s", task.ID.String())
			payloadFilePath = &proxyURL
		}
	}

	if payloadFilePath != nil || len(payloadFormats) > 0 {
		payload = &model.DownloadPayload{
			FilePath: payloadFilePath,
			Formats:  payloadFormats,
		}
	}

	log.Info().
		Str("task_id", task.ID.String()).
		Int("formats_count", len(payloadFormats)).
		Str("payload_file_path", func() string {
			if payloadFilePath != nil {
				return *payloadFilePath
			}
			return "nil"
		}()).
		Msg("Publishing completion event with payload")

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
