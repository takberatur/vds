package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/user/video-downloader-backend/internal/config"
)

var csrfInstance fiber.Handler

func CSRFMiddleware() fiber.Handler {
	if csrfInstance != nil {
		return csrfInstance
	}

	cfg := config.LoadConfig()
	isProd := cfg.AppEnv == "production"

	csrfInstance = csrf.New(csrf.Config{
		KeyLookup:      "header:X-XSRF-TOKEN", // Frontend sends this header
		CookieName:     "csrf_token",          // Cookie name to store the token
		CookieSameSite: "Lax",
		CookieSecure:   isProd,
		CookieHTTPOnly: false,
		ContextKey:     "csrf", // Explicitly set context key

		Expiration:   1 * time.Hour,
		KeyGenerator: utils.UUIDv4,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Invalid CSRF Token",
				"error":   err.Error(),
			})
		},
	})

	return csrfInstance
}
