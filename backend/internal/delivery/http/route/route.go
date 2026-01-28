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
		c.Redis,
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
	centrifugoHandler := handler.NewCentrifugoHandler(tokenService)
	_ = handler.NewSubscriptionHandler(subscriptionService)

	credentialLimiter := middleware.CredentialAttemptLimiter(c.Redis)
	csrfMiddleware := middleware.NewCSRF(c.Redis)

	api := c.App.Group("/api/v1")

	api.Get("/token/csrf", csrfMiddleware, func(c *fiber.Ctx) error {
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
	protectedAdmin.Put("/settings/bulk", csrfMiddleware, settingHandler.UpdateSettingsBulk)
	protectedAdmin.Post("/settings/upload", csrfMiddleware, settingHandler.UploadFile)

	// Admin users
	protectedAdmin.Get("/users/current", userHandler.GetCurrentUser)
	protectedAdmin.Put("/users/profile", csrfMiddleware, userHandler.UpdateProfile)
	protectedAdmin.Put("/users/password", csrfMiddleware, userHandler.UpdatePassword)
	protectedAdmin.Post("/users/avatar", csrfMiddleware, userHandler.UploadAvatar)

	// Dashboard
	protectedAdmin.Get("/dashboard", adminHandler.GetDashboardData)

	// Platforms (Added CRUD routes)
	protectedAdmin.Get("/platforms", platformHandler.GetPlatforms)
	protectedAdmin.Get("/platforms/:id", platformHandler.GetPlatformByID)
	protectedAdmin.Get("/platforms/type/:type", platformHandler.GetPlatformByType)
	protectedAdmin.Get("/platforms/slug/:slug", platformHandler.GetPlatformBySlug)
	protectedAdmin.Post("/platforms", csrfMiddleware, platformHandler.CreatePlatform)
	protectedAdmin.Put("/platforms/:id", csrfMiddleware, platformHandler.UpdatePlatform)
	protectedAdmin.Post("/platforms/thumbnail/:id", csrfMiddleware, platformHandler.UploadThumbnail)
	protectedAdmin.Delete("/platforms/:id", csrfMiddleware, platformHandler.DeletePlatform)
	protectedAdmin.Delete("/platforms/bulk", csrfMiddleware, platformHandler.BulkDeletePlatforms)

	// Application
	protectedAdmin.Get("/applications", applicationHandler.GetApplications)
	protectedAdmin.Get("/applications/:id", applicationHandler.FindByID)
	protectedAdmin.Post("/applications", csrfMiddleware, applicationHandler.RegisterApp)
	protectedAdmin.Delete("/applications/bulk", csrfMiddleware, applicationHandler.BulkDeleteApps)
	protectedAdmin.Put("/applications/:id", csrfMiddleware, applicationHandler.UpdateApp)
	protectedAdmin.Delete("/applications/:id", csrfMiddleware, applicationHandler.DeleteApp)

	// Downloads
	protectedAdmin.Get("/downloads", downloadHandler.GetDownloads)
	protectedAdmin.Get("/downloads/:id", downloadHandler.FindByID)
	protectedAdmin.Delete("/downloads/bulk", csrfMiddleware, downloadHandler.BulkDeleteDownloads)
	protectedAdmin.Put("/downloads/:id", csrfMiddleware, downloadHandler.UpdateDownload)
	protectedAdmin.Delete("/downloads/:id", csrfMiddleware, downloadHandler.DeleteDownload)

	// Health Check
	protectedAdmin.Get("/health/check", healthHandler.Check)
	protectedAdmin.Get("/health/log", healthHandler.GetLogger)
	protectedAdmin.Post("/health/log", csrfMiddleware, healthHandler.ClearLogs)

	// cookies
	protectedAdmin.Get("/cookies", adminHandler.GetCookies)
	protectedAdmin.Put("/cookies", csrfMiddleware, adminHandler.UpdateCookies)

	// Web Client Routes
	publicWeb.Get("/centrifugo/token", centrifugoHandler.GetToken)
	publicWeb.Post("/contact", webHandler.Contact)

	publicWeb.Get("/platforms", platformHandler.GetAll)
	publicWeb.Get("/platforms/:id", platformHandler.GetPlatformByID)
	publicWeb.Get("/platforms/type/:type", platformHandler.GetPlatformByType)
	publicWeb.Get("/platforms/slug/:slug", platformHandler.GetPlatformBySlug)
	publicWeb.Get("/platforms/category/:category", platformHandler.GetPlatformsByCategory)
	publicWeb.Post("/download/process/video", csrfMiddleware, downloadHandler.DownloadVideo)
	publicWeb.Post("/download/process/mp3", csrfMiddleware, downloadHandler.DownloadVideoToMp3)
	publicProxy.Get("/downloads/file/video", downloadHandler.ProxyDownload)
	publicProxy.Get("/downloads/file/mp3", downloadHandler.ProxyDownloadMp3)

	protectedUserWeb := publicWeb.Group("/protected-web", middleware.JWTMiddleware(tokenService))

	protectedUserWeb.Get("/users/current", userHandler.GetCurrentUser)
	protectedUserWeb.Put("/users/profile", csrfMiddleware, userHandler.UpdateProfile)
	protectedUserWeb.Put("/users/password", csrfMiddleware, userHandler.UpdatePassword)
	protectedUserWeb.Post("/users/avatar", csrfMiddleware, userHandler.UploadAvatar)
}
