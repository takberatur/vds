package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
)

type DownloadService interface {
	ProcessDownload(ctx context.Context, req model.DownloadRequest, userID *uuid.UUID, ip string) (*model.DownloadTask, error)
	GetUserHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.DownloadTask, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.DownloadTask, error)
	FindAll(ctx context.Context, params model.QueryParamsRequest) (*model.DownloadTasksResponse, error)
	Update(ctx context.Context, id uuid.UUID, task *model.DownloadTask) error
	Delete(ctx context.Context, id uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	GetTaskCookies(ctx context.Context, taskID uuid.UUID) (map[string]string, error)
}

type downloadService struct {
	repo         repository.DownloadRepository
	appRepo      repository.ApplicationRepository
	platformRepo repository.PlatformRepository
	downloader   infrastructure.DownloaderClient
	taskClient   infrastructure.TaskClient
	redisClient  *redis.Client
}

func NewDownloadService(
	repo repository.DownloadRepository,
	appRepo repository.ApplicationRepository,
	platformRepo repository.PlatformRepository,
	downloader infrastructure.DownloaderClient,
	taskClient infrastructure.TaskClient,
	redisClient *redis.Client,
) DownloadService {
	return &downloadService{
		repo:         repo,
		appRepo:      appRepo,
		platformRepo: platformRepo,
		downloader:   downloader,
		taskClient:   taskClient,
		redisClient:  redisClient,
	}
}

func (s *downloadService) ProcessDownload(ctx context.Context, req model.DownloadRequest, userID *uuid.UUID, ip string) (*model.DownloadTask, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 60*time.Second)
	defer cancel()

	var info *infrastructure.VideoInfo
	var formats []model.DownloadFormat
	var err error

	var appID *uuid.UUID
	if req.AppID != nil && *req.AppID != "" {
		id, err := uuid.Parse(*req.AppID)
		if err != nil {
			return nil, err
		}
		appPtr, err := s.appRepo.FindByID(subCtx, id)
		if err != nil {
			return nil, err
		}
		if appPtr == nil {
			return nil, errors.New("application not found")
		}
		if !appPtr.IsActive {
			return nil, errors.New("application is not active")
		}
		appID = &appPtr.ID
	}

	platform, err := s.platformRepo.FindByType(subCtx, req.Type)
	if err != nil {
		return nil, err
	}

	if !platform.IsActive {
		return nil, errors.New("platform is not active")
	}

	if typeAware, ok := s.downloader.(interface {
		GetVideoInfoWithType(ctx context.Context, url string, downloadType string) (*infrastructure.VideoInfo, error)
	}); ok {
		info, err = typeAware.GetVideoInfoWithType(subCtx, req.URL, platform.Type)
	} else {
		info, err = s.downloader.GetVideoInfo(subCtx, req.URL)
	}
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error().
				Err(err).
				Str("url", req.URL).
				Msg("Downloader timed out while fetching video info")
		}
		return nil, err
	}

	if info != nil && len(info.Formats) > 0 {
		// Filter formats logic
		var formatsToProcess []infrastructure.FormatInfo

		isTwitter := strings.ToLower(platform.Type) == "twitter" ||
			strings.ToLower(platform.Type) == "x" ||
			strings.Contains(strings.ToLower(req.URL), "twitter.com") ||
			strings.Contains(strings.ToLower(req.URL), "x.com") ||
			strings.Contains(strings.ToLower(req.URL), "twimg.com")

		isTiktok := strings.ToLower(platform.Type) == "tiktok" ||
			strings.Contains(strings.ToLower(req.URL), "tiktok.com")

		if strings.ToLower(platform.Type) == "youtube" ||
			strings.ToLower(platform.Type) == "facebook" ||
			isTwitter ||
			isTiktok {
			var validFormats []infrastructure.FormatInfo
			for _, f := range info.Formats {
				// Relaxed check: empty string often means "present but unknown"
				hasVideo := f.Vcodec != "none"
				hasAudio := f.Acodec != "none"

				if f.Vcodec == "" && f.Acodec == "" && f.Ext == "mp4" {
					hasVideo = true
					hasAudio = true
				}

				// Special case for TikTok
				if isTiktok && (strings.HasPrefix(f.URL, "http") && !strings.Contains(f.URL, ".m3u8")) {
					hasVideo = true
					hasAudio = true
				}

				if hasVideo && hasAudio {
					validFormats = append(validFormats, f)
				}
			}
			if len(validFormats) > 0 {
				formatsToProcess = validFormats
			} else {
				formatsToProcess = info.Formats
			}
		} else {
			formatsToProcess = info.Formats
		}

		/*
			// Update: We want ALL formats to be available for selection, especially now that Frontend supports "Video Only" / "Audio Only" badges.
			formatsToProcess = info.Formats
		*/

		formats = make([]model.DownloadFormat, 0, len(formatsToProcess))
		for _, f := range formatsToProcess {
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
	}

	format := "mp4"
	if req.Type == "youtube-to-mp3" {
		format = "mp3"
	}
	title := info.Title
	thumbnailURL := info.Thumbnail
	filePath := info.DownloadURL

	var duration *int
	if info.Duration != nil && *info.Duration > 0 {
		dur := int(*info.Duration)
		duration = &dur
	}

	fileSize := info.Filesize

	task := &model.DownloadTask{
		UserID:       userID,
		AppID:        appID,
		OriginalURL:  req.URL,
		PlatformID:   platform.ID,
		PlatformType: platform.Type,
		Status:       "queued",
		Title:        &title,
		ThumbnailURL: &thumbnailURL,
		Format:       &format,
		FilePath:     &filePath,
		Duration:     duration,
		FileSize:     fileSize,
		Formats:      formats,
		IPAddress:    &ip,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(subCtx, task); err != nil {
		return nil, err
	}

	if s.taskClient != nil {
		log.Info().
			Str("task_id", task.ID.String()).
			Str("url", task.OriginalURL).
			Msg("Enqueuing video download task")

		if err := s.taskClient.EnqueueVideoDownload(task); err != nil {
			log.Error().
				Err(err).
				Str("task_id", task.ID.String()).
				Msg("Failed to enqueue video download task")
			return nil, err
		}

		log.Info().
			Str("task_id", task.ID.String()).
			Msg("Successfully enqueued video download task")
	}

	return task, nil
}

func (s *downloadService) GetUserHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.DownloadTask, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.repo.FindByUserID(subCtx, userID, limit, offset)
}

func (s *downloadService) FindByID(ctx context.Context, id uuid.UUID) (*model.DownloadTask, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.FindByID(subCtx, id)
}

func (s *downloadService) FindAll(ctx context.Context, params model.QueryParamsRequest) (*model.DownloadTasksResponse, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	tasks, pagination, err := s.repo.FindAll(subCtx, params)
	if err != nil {
		return nil, err
	}
	return &model.DownloadTasksResponse{
		Data:       tasks,
		Pagination: pagination,
	}, nil
}

func (s *downloadService) Update(ctx context.Context, id uuid.UUID, task *model.DownloadTask) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	existing, err := s.repo.FindByID(subCtx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("download task not found")
	}
	task.ID = id
	return s.repo.Update(subCtx, task)
}

func (s *downloadService) Delete(ctx context.Context, id uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.Delete(subCtx, id)
}

func (s *downloadService) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.BulkDelete(subCtx, ids)
}

func (s *downloadService) GetTaskCookies(ctx context.Context, taskID uuid.UUID) (map[string]string, error) {
	key := "download:cookies:" + taskID.String()
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var cookies map[string]string
	if err := json.Unmarshal([]byte(val), &cookies); err != nil {
		return nil, err
	}
	return cookies, nil
}
