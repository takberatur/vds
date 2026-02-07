package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/dto"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/logger"
	"github.com/user/video-downloader-backend/pkg/response"
	"github.com/user/video-downloader-backend/pkg/utils"
)

type WebHandler struct {
	webService service.WebService
}

func NewWebHandler(webService service.WebService) *WebHandler {
	return &WebHandler{
		webService: webService,
	}
}

func (h *WebHandler) Contact(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req dto.ContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	if err := h.webService.Contact(ctx, &req); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to process request", err.Error())
	}

	return response.Success(c, "Contact message sent successfully", nil)
}
func (h *WebHandler) ReportError(c *fiber.Ctx) error {
	start := time.Now()

	var req dto.WebErrorReport
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	req.Message = strings.TrimSpace(req.Message)
	req.Error = strings.TrimSpace(req.Error)
	if strings.TrimSpace(req.IPAddress) == "" {
		req.IPAddress = c.IP()
	}
	if strings.TrimSpace(req.UserAgent) == "" {
		req.UserAgent = c.Get("User-Agent")
	}

	msgShort := truncate(req.Message, 500)
	errShort := truncate(req.Error, 1200)

	if errShort == "" {
		errShort = "-"
	}

	levelEvent := log.Error()
	if req.Level == "warn" || req.Level == "warning" {
		levelEvent = log.Warn()
	}

	levelEvent.
		Str("error", errShort).
		Str("message", msgShort).
		Str("platform_id", truncate(req.PlatformID, 30)).
		Str("ip_address", truncate(req.IPAddress, 20)).
		Str("user_agent", truncate(req.UserAgent, 40)).
		Str("url", truncate(req.URL, 200)).
		Str("method", truncate(req.Method, 10)).
		Str("request", truncate(req.Request, 200)).
		Int("status", req.Status).
		Str("level", truncate(req.Level, 10)).
		Str("locale", truncate(req.Locale, 20)).
		Str("user_id", truncate(req.UserID, 80)).
		Str("timestamp_ms", truncate(strconv.FormatInt(req.TimestampMs, 10), 40)).
		Msg("web error report")

	if req.Level == "error" || req.Level == "fatal" || req.Level == "crash" || req.Status >= 500 {
		logger.NotifyTelegram(
			"[web] %s platform_id:%s url:%s\nip_address:%s\nuser_agent:%s\nstatus: %d(%s) locale:%s\nmsg: %s\nerr: %s\ndur=%s",
			req.Level,
			req.PlatformID,
			truncate(req.URL, 200),
			truncate(req.IPAddress, 20),
			truncate(req.UserAgent, 40),
			req.Status,
			http.StatusText(req.Status),
			truncate(req.Locale, 20),
			msgShort,
			errShort,
			time.Since(start).Truncate(time.Millisecond).String(),
		)
	}

	return response.Success(c, "Error received", map[string]any{"ok": true})
}
