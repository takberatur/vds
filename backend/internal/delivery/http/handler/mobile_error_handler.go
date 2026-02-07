package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/pkg/logger"
	"github.com/user/video-downloader-backend/pkg/response"
)

type MobileErrorHandler struct{}

func NewMobileErrorHandler() *MobileErrorHandler {
	return &MobileErrorHandler{}
}

type mobileErrorPayload struct {
	Message     string            `json:"message"`
	Stack       string            `json:"stack"`
	Level       string            `json:"level"`
	Tag         string            `json:"tag"`
	AppVersion  string            `json:"app_version"`
	VersionCode string            `json:"version_code"`
	Android     string            `json:"android_version"`
	DeviceBrand string            `json:"device_brand"`
	DeviceModel string            `json:"device_model"`
	Device      string            `json:"device"`
	Abi         string            `json:"abi"`
	Locale      string            `json:"locale"`
	UserId      string            `json:"user_id"`
	Screen      string            `json:"screen"`
	TimestampMs  string `json:"timestamp_ms"`
	BuildType   string            `json:"build_type"`
	Extras      map[string]string `json:"extras"`
}

func (h *MobileErrorHandler) SendNotifError(c *fiber.Ctx) error {
	start := time.Now()
	var p mobileErrorPayload
	if err := c.BodyParser(&p); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	p.Message = strings.TrimSpace(p.Message)
	p.Stack = strings.TrimSpace(p.Stack)
	p.Level = strings.ToLower(strings.TrimSpace(p.Level))
	if p.Level == "" {
		p.Level = "error"
	}

	appID, _ := c.Locals("app_id").(uuid.UUID)
	platform, _ := c.Locals("platform").(string)
	actor := fmt.Sprintf("app_id=%s platform=%s session=%s", appID.String(), strings.TrimSpace(platform), maskToken(c.Get("X-Session-Id")))

	msgShort := truncate(p.Message, 500)
	errShort := truncate(p.Stack, 1200)
	if errShort == "" {
		errShort = "-"
	}

	levelEvent := log.Error()
	if p.Level == "warn" || p.Level == "warning" {
		levelEvent = log.Warn()
	}

	levelEvent.
		Str("source", "android").
		Str("actor", actor).
		Str("tag", truncate(p.Tag, 100)).
		Str("message", msgShort).
		Str("app_version", truncate(p.AppVersion, 40)).
		Str("version_code", truncate(p.VersionCode, 20)).
		Str("android_version", truncate(p.Android, 20)).
		Str("device_brand", truncate(p.DeviceBrand, 40)).
		Str("device_model", truncate(p.DeviceModel, 80)).
		Str("device", truncate(p.Device, 80)).
		Str("abi", truncate(p.Abi, 40)).
		Str("locale", truncate(p.Locale, 20)).
		Str("user_id", truncate(p.UserId, 80)).
		Str("screen", truncate(p.Screen, 120)).
		Str("timestamp_ms", truncate(p.TimestampMs, 40)).
		Str("build_type", truncate(p.BuildType, 30)).
		Msg("android_error_report")

	if p.Level == "error" || p.Level == "fatal" || p.Level == "crash" {
		logger.NotifyTelegram(
			"[android] %s actor=%s tag=%s\nmsg: %s\napp: %s(%s) android:%s\ndevice: %s %s\nscreen: %s\nstack: %s\ndur=%s",
			p.Level,
			actor,
			truncate(p.Tag, 60),
			msgShort,
			truncate(p.AppVersion, 40),
			truncate(p.VersionCode, 20),
			truncate(p.Android, 20),
			truncate(p.DeviceBrand, 30),
			truncate(p.DeviceModel, 40),
			truncate(p.Screen, 100),
			errShort,
			time.Since(start).Truncate(time.Millisecond).String(),
		)
	}

	return response.Success(c, "Error received", map[string]any{"ok": true})
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func maskToken(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	if len(s) <= 10 {
		if len(s) <= 3 {
			return s + "..."
		}
		return s[:3] + "..."
	}
	return s[:6] + "..." + s[len(s)-4:]
}
