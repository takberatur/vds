package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/user/video-downloader-backend/internal/config"
)

func NewCSRF(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cfg := config.LoadConfig()
		isProd := cfg.AppEnv == "production"

		// Cookie name for the session ID
		cookieName := "csrf_session_id"

		// Determine Cookie Settings
		// If production, use None (requires Secure). If dev, use Lax.
		sameSite := "Lax"
		if isProd {
			sameSite = "None"
		}

		// 1. Get or Create Session ID
		sessionID := c.Cookies(cookieName)
		if sessionID == "" {
			sessionID = uuid.New().String()
			c.Cookie(&fiber.Cookie{
				Name:     cookieName,
				Value:    sessionID,
				Expires:  time.Now().Add(24 * time.Hour),
				HTTPOnly: true,
				Secure:   isProd,
				SameSite: sameSite,
			})
		}

		// Redis Key
		redisKey := fmt.Sprintf("csrf_token:%s", sessionID)

		// 2. Handle GET (Token Generation)
		if c.Method() == fiber.MethodGet {
			// Check if token exists to reuse (better for multi-tab support)
			storedToken, err := redisClient.Get(c.Context(), redisKey).Result()
			if err == nil && storedToken != "" {
				// Extend TTL
				redisClient.Expire(c.Context(), redisKey, 1*time.Hour)
				c.Locals("csrf", storedToken)
				return c.Next()
			}

			// Generate new token
			token := uuid.New().String()

			// Store in Redis with 1 hour expiration
			err = redisClient.Set(c.Context(), redisKey, token, 1*time.Hour).Err()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Failed to generate CSRF token",
				})
			}

			// Set in Locals for the handler to return
			c.Locals("csrf", token)
			return c.Next()
		}

		// 3. Handle Mutation Methods (Validation)
		// Get token from header
		clientToken := c.Get("X-XSRF-TOKEN")
		if clientToken == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "CSRF Token missing",
			})
		}

		// Get token from Redis
		storedToken, err := redisClient.Get(c.Context(), redisKey).Result()
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired CSRF session",
			})
		}

		if clientToken != storedToken {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Invalid CSRF Token",
			})
		}

		return c.Next()
	}
}
