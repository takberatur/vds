package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/redis/go-redis/v9"
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

var rateLimitDownloadLua = redis.NewScript(`
local key = KEYS[1]
local max = tonumber(ARGV[1])
local exp = tonumber(ARGV[2])

local current = redis.call('INCR', key)
if current == 1 then
  redis.call('EXPIRE', key, exp)
end

local ttl = redis.call('TTL', key)
return {current, ttl}
`)

func RateLimitDownloadRedis(rdb *redis.Client) fiber.Handler {
	max := int64(3)
	expSeconds := int64(60)

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if ip == "" {
			ip = "unknown"
		}

		key := "rl:download:" + ip
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		res, err := rateLimitDownloadLua.Run(ctx, rdb, []string{key}, max, expSeconds).Result()
		if err != nil {
			return response.Error(c, fiber.StatusTooManyRequests, "Too many requests, please try again later after 1 minute", nil)
		}

		vals, ok := res.([]any)
		if !ok || len(vals) < 2 {
			return response.Error(c, fiber.StatusTooManyRequests, "Too many requests, please try again later after 1 minute", nil)
		}

		var current int64
		switch v := vals[0].(type) {
		case int64:
			current = v
		case string:
			current, _ = strconv.ParseInt(v, 10, 64)
		}

		var ttl int64
		switch v := vals[1].(type) {
		case int64:
			ttl = v
		case string:
			ttl, _ = strconv.ParseInt(v, 10, 64)
		}
		if ttl < 0 {
			ttl = expSeconds
		}

		if current > max {
			c.Set("Retry-After", strconv.FormatInt(ttl, 10))
			return response.Error(c, fiber.StatusTooManyRequests, "Too many requests, please try again later after 1 minute", nil)
		}

		return c.Next()
	}
}
