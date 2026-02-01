package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/user/video-downloader-backend/pkg/response"
)

func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,             // Max 100 requests
		Expiration: 1 * time.Minute, // Per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusTooManyRequests, "Too many requests, please try again later after 1 minute", nil)
		},
	})
}
func RateLimitDownload() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,               // Max 3 requests
		Expiration: 1 * time.Minute, // Per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusTooManyRequests, "Too many requests, please try again later after 1 minute", nil)
		},
	})
}
