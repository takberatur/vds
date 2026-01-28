package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
)

type PlatformService interface {
	GetAll(ctx context.Context) ([]model.Platform, error)
	GetPlatforms(ctx context.Context, params model.QueryParamsRequest) (*model.PlatformsResponse, error)
	GetPlatformByID(ctx context.Context, id uuid.UUID) (*model.Platform, error)
	GetPlatformBySlug(ctx context.Context, slug string) (*model.Platform, error)
	GetPlatformByType(ctx context.Context, type_ string) (*model.Platform, error)
	GetPlatformsByCategory(ctx context.Context, category string) ([]model.Platform, error)
	CreatePlatform(ctx context.Context, platform *model.Platform) error
	UpdatePlatform(ctx context.Context, platform *model.Platform) error
	UploadThumbnail(ctx context.Context, platformID uuid.UUID, file io.Reader, filename string, size int64, contentType string) (string, error)
	DeletePlatform(ctx context.Context, id uuid.UUID) error
	BulkDeletePlatforms(ctx context.Context, ids []uuid.UUID) error
}

type platformService struct {
	repo          repository.PlatformRepository
	storageClient infrastructure.StorageClient
	cfg           *config.Config
}

func NewPlatformService(repo repository.PlatformRepository, storageClient infrastructure.StorageClient, cfg *config.Config) PlatformService {
	return &platformService{repo: repo, storageClient: storageClient, cfg: cfg}
}

func (s *platformService) GetAll(ctx context.Context) ([]model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.GetAll(subCtx)
}

func (s *platformService) GetPlatforms(ctx context.Context, params model.QueryParamsRequest) (*model.PlatformsResponse, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()
	data, pagination, err := s.repo.FindAll(subCtx, params)
	if err != nil {
		return nil, err
	}

	return &model.PlatformsResponse{
		Data:       data,
		Pagination: pagination,
	}, nil
}

func (s *platformService) GetPlatformByID(ctx context.Context, id uuid.UUID) (*model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.FindByID(subCtx, id)
}

func (s *platformService) GetPlatformBySlug(ctx context.Context, slug string) (*model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.FindBySlug(subCtx, slug)
}

func (s *platformService) GetPlatformByType(ctx context.Context, type_ string) (*model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.FindByType(subCtx, type_)
}

func (s *platformService) GetPlatformsByCategory(ctx context.Context, category string) ([]model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.FindByCategory(subCtx, category)
}

func (s *platformService) CreatePlatform(ctx context.Context, platform *model.Platform) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()
	return s.repo.Create(subCtx, platform)
}

func (s *platformService) UpdatePlatform(ctx context.Context, platform *model.Platform) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()
	return s.repo.Update(subCtx, platform)
}

func (s *platformService) UploadThumbnail(ctx context.Context, platformID uuid.UUID, file io.Reader, filename string, size int64, contentType string) (string, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	currentPlatform, err := s.repo.FindByID(subCtx, platformID)
	if err == nil && currentPlatform != nil && currentPlatform.ThumbnailURL != "" {
		oldThumbnailURL := currentPlatform.ThumbnailURL

		parsedURL, err := url.Parse(oldThumbnailURL)
		if err == nil {
			path := parsedURL.Path
			path = strings.TrimPrefix(path, "/")

			if strings.HasPrefix(path, s.cfg.MinioBucket+"/") {
				objectName := strings.TrimPrefix(path, s.cfg.MinioBucket+"/")

				log.Info().Str("platformID", platformID.String()).Str("object", objectName).Msg("Deleting old thumbnail")
				if err := s.storageClient.DeleteFile(subCtx, s.cfg.MinioBucket, objectName); err != nil {
					log.Error().Err(err).Str("object", objectName).Msg("Failed to delete old avatar")
				}
			}
		}
	}

	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("thumbnails/%s_%d%s", platformID.String(), time.Now().Unix(), ext)

	uploadedPath, err := s.storageClient.UploadFile(subCtx, s.cfg.MinioBucket, objectName, file, size, contentType)
	if err != nil {
		return "", fmt.Errorf("storage upload failed: %w", err)
	}

	if err := s.repo.UpdateThumbnail(subCtx, uploadedPath, platformID); err != nil {
		return "", fmt.Errorf("db update failed: %w", err)
	}

	presignedURL, err := s.storageClient.GetFileURL(subCtx, s.cfg.MinioBucket, uploadedPath, 7*24*time.Hour)
	if err != nil {
		return uploadedPath, nil
	}

	return presignedURL, nil
}

func (s *platformService) DeletePlatform(ctx context.Context, id uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.Delete(subCtx, id)
}

func (s *platformService) BulkDeletePlatforms(ctx context.Context, ids []uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.BulkDelete(subCtx, ids)
}
