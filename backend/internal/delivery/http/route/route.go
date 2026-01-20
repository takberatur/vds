package route

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/delivery/helpers"
	"github.com/user/video-downloader-backend/internal/delivery/http/handler"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/repository"
	"github.com/user/video-downloader-backend/internal/service"
)

type RouteConfig struct {
	App           *fiber.App
	DB            *infrastructure.Database
	Redis         *redis.Client
	Cfg           *config.Config
	StorageClient infrastructure.StorageClient
}

func SetupRoutes(c *RouteConfig) {
	settingRepo := repository.NewSettingRepository(c.DB.Pool)
	userRepo := repository.NewUserRepository(c.DB.Pool)
	platformRepo := repository.NewPlatformRepository(c.DB.Pool) // Added Platform Repo
	adminRepo := repository.NewAdminRepository(c.DB.Pool)
	applicationRepo := repository.NewApplicationRepository(c.DB.Pool)
	downloadRepo := repository.NewDownloadRepository(c.DB.Pool)
	subscriptionRepo := repository.NewSubscriptionRepository(c.DB.Pool)

	tokenService := service.NewTokenService(c.Cfg)
	mailHelper := helpers.NewMailHelper(settingRepo)
	authService := service.NewAuthService(userRepo, mailHelper, tokenService, c.Redis)

	settingService := service.NewSettingService(settingRepo, c.StorageClient)
	userService := service.NewUserService(userRepo, c.StorageClient, c.Cfg)
	platformService := service.NewPlatformService(platformRepo, c.StorageClient, c.Cfg)
	adminService := service.NewAdminService(adminRepo)
	applicationService := service.NewApplicationService(applicationRepo)
	webService := service.NewWebService(mailHelper)

	downloader := infrastructure.NewFallbackDownloader()
	taskClient := infrastructure.NewTaskClient(c.Cfg.RedisAddr, c.Cfg.RedisPassword)
	downloadService := service.NewDownloadService(
		downloadRepo,
		applicationRepo,
		platformRepo,
		downloader,
		taskClient,
	)

	subscriptionService := service.NewSubscriptionService(subscriptionRepo)

	// Handlers
	healthHandler := handler.NewHealthHandler(c.DB.Pool, c.Redis)
	authHandler := handler.NewAuthHandler(authService)
	settingHandler := handler.NewSettingHandler(settingService)
	userHandler := handler.NewUserHandler(userService)
	platformHandler := handler.NewPlatformHandler(platformService) // Added Platform
	adminHandler := handler.NewAdminHandler(adminService)
	applicationHandler := handler.NewApplicationHandler(applicationService)
	downloadHandler := handler.NewDownloadHandler(downloadService, userService)
	webHandler := handler.NewWebHandler(webService)
	_ = handler.NewSubscriptionHandler(subscriptionService)

	credentialLimiter := middleware.CredentialAttemptLimiter(c.Redis)

	api := c.App.Group("/api/v1")

	api.Get("/token/csrf", middleware.CSRFMiddleware(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"csrf_token": c.Locals("csrf"),
			},
		})
	})

	api.Get("/ws", websocket.New(downloadHandler.DownloadEvents))
	api.Get("/downloads/ws/:user_id", websocket.New(downloadHandler.DownloadEventsByUser))

	api.Use(middleware.SetTimeoutContext(60 * time.Second))

	publicAdmin := api.Group("/public-admin")
	publicWeb := api.Group("/web-client")
	publicProxy := api.Group("/public-proxy")
	// Future route for android app client
	// publicMobile := api.Group("/mobile-client")

	publicAdmin.Post("/auth/google", credentialLimiter, authHandler.GoogleLogin)
	publicAdmin.Post("/auth/email", credentialLimiter, authHandler.LoginEmail)
	publicAdmin.Post("/auth/forgot-password", credentialLimiter, authHandler.ForgotPassword)
	publicAdmin.Post("/auth/reset-password", credentialLimiter, authHandler.ResetPassword)
	publicAdmin.Post("/auth/logout", authHandler.Logout)

	publicAdmin.Get("/settings/public", settingHandler.GetPublicSettings)

	// Protected Admin Routes
	protectedAdmin := api.Group("/protected-admin", middleware.JWTMiddleware(tokenService), middleware.AdminMiddleware())

	protectedAdmin.Post("/auth/logout", authHandler.Logout)

	// Settings
	protectedAdmin.Get("/settings", settingHandler.GetAllSettings)
	protectedAdmin.Put("/settings/bulk", middleware.CSRFMiddleware(), settingHandler.UpdateSettingsBulk)
	protectedAdmin.Post("/settings/upload", middleware.CSRFMiddleware(), settingHandler.UploadFile)

	// Admin users
	protectedAdmin.Get("/users/current", userHandler.GetCurrentUser)
	protectedAdmin.Put("/users/profile", middleware.CSRFMiddleware(), userHandler.UpdateProfile)
	protectedAdmin.Put("/users/password", middleware.CSRFMiddleware(), userHandler.UpdatePassword)
	protectedAdmin.Post("/users/avatar", middleware.CSRFMiddleware(), userHandler.UploadAvatar)

	// Dashboard
	protectedAdmin.Get("/dashboard", adminHandler.GetDashboardData)

	// Platforms (Added CRUD routes)
	protectedAdmin.Get("/platforms", platformHandler.GetPlatforms)
	protectedAdmin.Get("/platforms/:id", platformHandler.GetPlatformByID)
	protectedAdmin.Get("/platforms/type/:type", platformHandler.GetPlatformByType)
	protectedAdmin.Get("/platforms/slug/:slug", platformHandler.GetPlatformBySlug)
	protectedAdmin.Post("/platforms", middleware.CSRFMiddleware(), platformHandler.CreatePlatform)
	protectedAdmin.Put("/platforms/:id", middleware.CSRFMiddleware(), platformHandler.UpdatePlatform)
	protectedAdmin.Post("/platforms/thumbnail/:id", middleware.CSRFMiddleware(), platformHandler.UploadThumbnail)
	protectedAdmin.Delete("/platforms/:id", middleware.CSRFMiddleware(), platformHandler.DeletePlatform)
	protectedAdmin.Delete("/platforms/bulk", middleware.CSRFMiddleware(), platformHandler.BulkDeletePlatforms)

	// Application
	protectedAdmin.Get("/applications", applicationHandler.GetApplications)
	protectedAdmin.Get("/applications/:id", applicationHandler.FindByID)
	protectedAdmin.Post("/applications", middleware.CSRFMiddleware(), applicationHandler.RegisterApp)
	protectedAdmin.Delete("/applications/bulk", middleware.CSRFMiddleware(), applicationHandler.BulkDeleteApps)
	protectedAdmin.Put("/applications/:id", middleware.CSRFMiddleware(), applicationHandler.UpdateApp)
	protectedAdmin.Delete("/applications/:id", middleware.CSRFMiddleware(), applicationHandler.DeleteApp)

	// Downloads
	protectedAdmin.Get("/downloads", downloadHandler.GetDownloads)
	protectedAdmin.Get("/downloads/:id", downloadHandler.FindByID)
	protectedAdmin.Delete("/downloads/bulk", middleware.CSRFMiddleware(), downloadHandler.BulkDeleteDownloads)
	protectedAdmin.Put("/downloads/:id", middleware.CSRFMiddleware(), downloadHandler.UpdateDownload)
	protectedAdmin.Delete("/downloads/:id", middleware.CSRFMiddleware(), downloadHandler.DeleteDownload)

	// Health Check
	protectedAdmin.Get("/health/check", healthHandler.Check)
	protectedAdmin.Get("/health/log", healthHandler.GetLogger)
	protectedAdmin.Post("/health/log", middleware.CSRFMiddleware(), healthHandler.ClearLogs)

	// Web Client Routes
	publicWeb.Post("/contact", webHandler.Contact)

	publicWeb.Get("/platforms", platformHandler.GetAll)
	publicWeb.Get("/platforms/:id", platformHandler.GetPlatformByID)
	publicWeb.Get("/platforms/type/:type", platformHandler.GetPlatformByType)
	publicWeb.Get("/platforms/slug/:slug", platformHandler.GetPlatformBySlug)
	publicWeb.Post("/download/process", downloadHandler.DownloadVideo)
	publicProxy.Get("/downloads/file", downloadHandler.ProxyDownload)

	protectedUserWeb := publicWeb.Group("/protected-web", middleware.JWTMiddleware(tokenService))

	protectedUserWeb.Get("/users/current", userHandler.GetCurrentUser)
	protectedUserWeb.Put("/users/profile", middleware.CSRFMiddleware(), userHandler.UpdateProfile)
	protectedUserWeb.Put("/users/password", middleware.CSRFMiddleware(), userHandler.UpdatePassword)
	protectedUserWeb.Post("/users/avatar", middleware.CSRFMiddleware(), userHandler.UploadAvatar)
}
