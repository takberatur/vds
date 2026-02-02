package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/pkg/response"
)

type BootstrapHandler struct {
	redis *redis.Client
}

func NewBootstrapHandler(redis *redis.Client) *BootstrapHandler {
	return &BootstrapHandler{redis: redis}
}

type mobileSessionRecord struct {
	Secret   string    `json:"secret"`
	AppID    uuid.UUID `json:"app_id"`
	Platform string    `json:"platform"`
	ExpiresAt time.Time `json:"expires_at"`
}

type bootstrapResponse struct {
	SessionID     string `json:"session_id"`
	SessionSecret string `json:"session_secret"`
	ExpiresInSec  int64  `json:"expires_in"`
}

func (h *BootstrapHandler) Bootstrap(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	appID, ok := c.Locals("app_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}
	platform, _ := c.Locals("platform").(string)

	sessionID, err := randomToken(24)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create session", err.Error())
	}
	secret, err := randomToken(48)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create session", err.Error())
	}

	expiresIn := int64(10 * 60)
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	rec := mobileSessionRecord{
		Secret:   secret,
		AppID:    appID,
		Platform: platform,
		ExpiresAt: expiresAt,
	}
	buf, err := json.Marshal(rec)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create session", err.Error())
	}

	key := "mobile_session:" + sessionID
	if err := h.redis.Set(ctx, key, string(buf), time.Duration(expiresIn)*time.Second).Err(); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create session", err.Error())
	}

	return response.Success(c, "Bootstrap successful", bootstrapResponse{
		SessionID:     sessionID,
		SessionSecret: secret,
		ExpiresInSec:  expiresIn,
	})
}

func randomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

