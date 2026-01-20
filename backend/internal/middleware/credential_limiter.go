package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/user/video-downloader-backend/pkg/response"
)

type CredentialLimiterConfig struct {
	MaxAttempts int
	Window      time.Duration
	BlockTime   time.Duration
}

func CredentialAttemptLimiter(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := FromContext(c)

		key := fmt.Sprintf("attempt:%s:%s", c.Path(), c.IP())

		attempts, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return c.Next()
		}

		if attempts >= 3 {
			ttl, _ := redisClient.TTL(ctx, key).Result()
			return response.Error(c, fiber.StatusTooManyRequests, fmt.Sprintf("Too many failed attempts. Please try again in %v", ttl), nil)
		}

		err = c.Next()

		if c.Response().StatusCode() == fiber.StatusUnauthorized || c.Response().StatusCode() == fiber.StatusBadRequest {
			pipe := redisClient.Pipeline()
			pipe.Incr(ctx, key)
			if attempts == 0 {
				pipe.Expire(ctx, key, 15*time.Minute)
			}
			_, _ = pipe.Exec(ctx)
		} else if c.Response().StatusCode() == fiber.StatusOK {
			redisClient.Del(ctx, key)
		}

		return err
	}
}
