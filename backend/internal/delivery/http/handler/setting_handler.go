package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
)

type SettingHandler struct {
	svc service.SettingService
}

func NewSettingHandler(svc service.SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

func (h *SettingHandler) UpdateSettingsBulk(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
	scope := c.Query("scope")
	if scope == "" {
		scope = middleware.GetSettingsScope(c)
	}

	var settings []model.UpdateSettingsBulkRequest
	if err := c.BodyParser(&settings); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.svc.UpdateSettingsBulk(ctx, scope, settings); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update settings", err.Error())
	}

	return response.Success(c, "Settings updated successfully", nil)
}

func (h *SettingHandler) UploadFile(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
	scope := c.Query("scope")
	if scope == "" {
		scope = middleware.GetSettingsScope(c)
	}

	req := new(model.UploadFileRequest)
	if err := c.BodyParser(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.Key == "" {
		req.Key = c.FormValue("key")
	}

	if req.Key != "site_logo" && req.Key != "site_favicon" {
		return response.Error(c, fiber.StatusBadRequest, "Invalid key. Must be 'site_logo' or 'site_favicon'", nil)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "File is required", err.Error())
	}

	url, err := h.svc.UploadFile(ctx, scope, file, req.Key)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to upload file", err.Error())
	}

	return response.Success(c, "File uploaded successfully", fiber.Map{"url": url})
}

func (h *SettingHandler) GetPublicSettings(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
	scope := c.Query("scope")
	if scope == "" {
		scope = middleware.GetSettingsScope(c)
	}

	settings, err := h.svc.GetPublicSettings(ctx, scope)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch settings", err.Error())
	}
	return response.Success(c, "Settings fetched successfully", settings)
}

func (h *SettingHandler) GetAllSettings(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
	scope := c.Query("scope")
	if scope == "" {
		scope = middleware.GetSettingsScope(c)
	}

	settings, err := h.svc.GetAllSettings(ctx, scope)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch settings", err.Error())
	}
	return response.Success(c, "All settings fetched", settings)
}
