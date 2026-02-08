package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
)

func APIKeyMiddleware(appCache service.AppCacheService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := FromContext(c)

		path := c.Path()
		if strings.HasPrefix(path, "/.well-known/") {
			return c.Next()
		}

		if path == "/metrics" || strings.HasPrefix(path, "/api/v1/public-admin") || strings.HasPrefix(path, "/api/v1/protected-admin") || strings.HasPrefix(path, "/api/v1/token/csrf") || strings.HasPrefix(path, "/api/v1/web-client") || strings.HasPrefix(path, "/api/v1/public-proxy") || strings.HasPrefix(path, "/api/v1/downloads/ws") || strings.HasPrefix(path, "/api/v1/ws") {
			return c.Next()
		}

		if strings.HasPrefix(path, "/api/v1/mobile-client") && path != "/api/v1/mobile-client/bootstrap" {
			if c.Get("X-Session-Id") != "" {
				return c.Next()
			}
		}

		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return response.Error(c, fiber.StatusUnauthorized, "API Key is missing", nil)
		}

		app, err := appCache.GetAppByAPIKey(ctx, apiKey)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to validate API Key", nil)
		}

		if app == nil || !app.IsActive {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid or inactive API Key", nil)
		}

		c.Locals("app_id", app.ID)
		c.Locals("platform", app.Platform)

		return c.Next()
	}
}
