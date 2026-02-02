package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/user/video-downloader-backend/pkg/response"
)

type mobileSessionRecord struct {
	Secret   string    `json:"secret"`
	AppID    uuid.UUID `json:"app_id"`
	Platform string    `json:"platform"`
	ExpiresAt time.Time `json:"expires_at"`
}

func MobileSignatureMiddleware(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if !strings.HasPrefix(path, "/api/v1/mobile-client") {
			return c.Next()
		}
		if path == "/api/v1/mobile-client/bootstrap" {
			return c.Next()
		}

		sessionID := c.Get("X-Session-Id")
		timestampStr := c.Get("X-Timestamp")
		nonce := c.Get("X-Nonce")
		signature := c.Get("X-Signature")
		if sessionID == "" || timestampStr == "" || nonce == "" || signature == "" {
			return response.Error(c, fiber.StatusUnauthorized, "Missing signature headers", nil)
		}

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid timestamp", nil)
		}
		now := time.Now().Unix()
		if timestamp < now-300 || timestamp > now+300 {
			return response.Error(c, fiber.StatusUnauthorized, "Timestamp out of range", nil)
		}

		ctx := HandlerContext(c)
		val, err := redisClient.Get(ctx, "mobile_session:"+sessionID).Result()
		if err == redis.Nil {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid session", nil)
		}
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to validate session", err.Error())
		}

		var rec mobileSessionRecord
		if err := json.Unmarshal([]byte(val), &rec); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to validate session", err.Error())
		}

		nonceKey := fmt.Sprintf("mobile_nonce:%s:%s", sessionID, nonce)
		ok, err := redisClient.SetNX(ctx, nonceKey, "1", 5*time.Minute).Result()
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to validate request", err.Error())
		}
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, "Replay detected", nil)
		}

		bodyHash := sha256Hex(c.Body())
		url := c.OriginalURL()
		canonical := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", c.Method(), url, timestampStr, nonce, bodyHash)
		expected := hmacSHA256Hex(rec.Secret, canonical)
		if !hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(signature))) {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid signature", nil)
		}

		c.Locals("app_id", rec.AppID)
		c.Locals("platform", rec.Platform)

		return c.Next()
	}
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hmacSHA256Hex(secret string, msg string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

