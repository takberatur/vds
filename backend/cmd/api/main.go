package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/delivery/http/handler"
	"github.com/user/video-downloader-backend/internal/delivery/http/route"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/logger"
)

func main() {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger.Init()

	log.Info().Msg("Starting application...")

	cfg := config.LoadConfig()

	db, err := infrastructure.NewPostgresClient(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Pool.Close()
	log.Info().Msg("Connected to PostgreSQL")

		redisClient, err := infrastructure.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to redis")
	} else {
		defer redisClient.Close()
		log.Info().Msg("Connected to Redis")

		go func() {
				subCtx := context.Background()
				sub := redisClient.Subscribe(subCtx, infrastructure.DownloadEventChannel)
			ch := sub.Channel()
			log.Info().Str("channel", infrastructure.DownloadEventChannel).Msg("Subscribed to download events channel")
			for msg := range ch {
				var event model.DownloadEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal download event from redis")
					continue
				}
				log.Info().
					Str("type", event.Type).
					Str("task_id", event.TaskID.String()).
					Str("status", event.Status).
					Msg("Received download event from redis, broadcasting to websockets")
				handler.BroadcastDownloadEvent(&event)
			}
			log.Warn().Msg("Redis download events subscription loop exited")
		}()
	}

	storageClient, err := infrastructure.NewStorageClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioUseSSL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to minio")
	} else {
		log.Info().Msg("Connected to MinIO")
		if err := storageClient.CreateBucket(ctxTimeout, cfg.MinioBucket); err != nil {
			log.Error().Err(err).Str("bucket", cfg.MinioBucket).Msg("Failed to create bucket")
		}
	}

	applicationRepo := repository.NewApplicationRepository(db.Pool)
	appCacheService := service.NewAppCacheService(applicationRepo, redisClient)

	if err := appCacheService.LoadAllAppsToCache(ctxTimeout); err != nil {
		log.Error().Err(err).Msg("Failed to load apps to cache")
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	infrastructure.SetupMetrics(app)

	app.Use(
		middleware.ContextMiddleware(30*time.Minute),
		recover.New(),
		middleware.RequestLogger(),
		middleware.RateLimiter(),
		cors.New(cors.Config{
			AllowOrigins:     cfg.ClientURL,
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-Key, X-XSRF-TOKEN",
			AllowCredentials: true,
		}),
	)
	app.Use(middleware.APIKeyMiddleware(appCacheService))

	routeConfig := &route.RouteConfig{
		App:           app,
		DB:            db,
		Redis:         redisClient,
		Cfg:           cfg,
		StorageClient: storageClient,
	}
	route.SetupRoutes(routeConfig)

	log.Info().Str("port", cfg.AppPort).Msg("Server starting")
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Error().Err(err).Msg("Server failed to start")
	}
}
