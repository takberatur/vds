package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
)

const (
	AppCacheKeyPrefix = "app:apikey:"
	AppCacheTTL       = 24 * time.Hour
)

type AppCacheService interface {
	LoadAllAppsToCache(ctx context.Context) error
	CacheApp(ctx context.Context, app *model.Application) error
	GetAppByAPIKey(ctx context.Context, apiKey string) (*model.Application, error)
	InvalidateApp(ctx context.Context, apiKey string) error
}

type appCacheService struct {
	repo  repository.ApplicationRepository
	redis *redis.Client
}

func NewAppCacheService(repo repository.ApplicationRepository, redis *redis.Client) AppCacheService {
	return &appCacheService{
		repo:  repo,
		redis: redis,
	}
}

func (s *appCacheService) LoadAllAppsToCache(ctx context.Context) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	log.Info().Msg("Starting to load all applications to Redis cache...")

	apps, err := s.repo.GetAll(subCtx)
	if err != nil {
		return err
	}

	count := 0
	pipeline := s.redis.Pipeline()
	for _, app := range apps {
		if !app.IsActive {
			continue
		}
		data, err := json.Marshal(app)
		if err != nil {
			log.Error().Err(err).Str("id", app.ID.String()).Msg("Failed to marshal app")
			continue
		}
		pipeline.Set(subCtx, AppCacheKeyPrefix+app.APIKey, data, AppCacheTTL)
		count++
	}

	_, err = pipeline.Exec(subCtx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	log.Info().Int("count", count).Msg("Successfully loaded applications to cache")
	return nil
}

func (s *appCacheService) CacheApp(ctx context.Context, app *model.Application) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	data, err := json.Marshal(app)
	if err != nil {
		return err
	}
	return s.redis.Set(subCtx, AppCacheKeyPrefix+app.APIKey, data, AppCacheTTL).Err()
}

func (s *appCacheService) GetAppByAPIKey(ctx context.Context, apiKey string) (*model.Application, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	data, err := s.redis.Get(subCtx, AppCacheKeyPrefix+apiKey).Bytes()
	if err == nil {
		var app model.Application
		if err := json.Unmarshal(data, &app); err == nil {
			return &app, nil
		}
	}

	app, err := s.repo.FindByAPIKey(subCtx, apiKey)
	if err != nil {
		return nil, err
	}
	if app != nil {
		_ = s.CacheApp(subCtx, app)
	}

	return app, nil
}

func (s *appCacheService) InvalidateApp(ctx context.Context, apiKey string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.redis.Del(subCtx, AppCacheKeyPrefix+apiKey).Err()
}
