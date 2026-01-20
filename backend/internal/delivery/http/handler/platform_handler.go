package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
)

type PlatformHandler struct {
	service service.PlatformService
}

func NewPlatformHandler(service service.PlatformService) *PlatformHandler {
	return &PlatformHandler{service: service}
}

func (h *PlatformHandler) GetAll(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	platforms, err := h.service.GetAll(ctx)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch platforms", err.Error())
	}
	return response.Success(c, "Platforms retrieved successfully", platforms)
}

func (h *PlatformHandler) GetPlatforms(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	// Parse Query Params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	params := model.QueryParamsRequest{
		Search:  c.Query("search"),
		SortBy:  c.Query("sort_by", "created_at"),
		OrderBy: c.Query("order_by", "desc"),
		Page:    page,
		Limit:   limit,
		Status:  c.Query("status"),
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			params.DateFrom = t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			params.DateTo = t
		}
	}

	resp, err := h.service.GetPlatforms(ctx, params)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch platforms", err.Error())
	}

	return response.SuccessWithMeta(c, "Platforms retrieved successfully",
		resp.Data,
		resp.Pagination,
	)

}

func (h *PlatformHandler) GetPlatformByID(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "ID is required", nil)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID format", err.Error())
	}

	platform, err := h.service.GetPlatformByID(ctx, id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch platform", err.Error())
	}

	return response.Success(c, "Platform retrieved successfully", platform)
}

func (h *PlatformHandler) GetPlatformBySlug(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	slug := c.Params("slug")
	if slug == "" {
		return response.Error(c, fiber.StatusBadRequest, "Slug is required", nil)
	}

	platform, err := h.service.GetPlatformBySlug(ctx, slug)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch platform", err.Error())
	}

	return response.Success(c, "Platform retrieved successfully", platform)
}

func (h *PlatformHandler) CreatePlatform(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var platform model.Platform
	if err := c.BodyParser(&platform); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.service.CreatePlatform(ctx, &platform); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create platform", err.Error())
	}

	return response.Success(c, "Platform created successfully", platform)
}

func (h *PlatformHandler) UpdatePlatform(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "ID is required", nil)
	}

	var platform model.Platform
	if err := c.BodyParser(&platform); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID format", err.Error())
	}
	platform.ID = id

	if err := h.service.UpdatePlatform(ctx, &platform); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update platform", err.Error())
	}

	return response.Success(c, "Platform updated successfully", platform)
}

func (h *PlatformHandler) DeletePlatform(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "ID is required", nil)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID format", err.Error())
	}

	if err := h.service.DeletePlatform(ctx, id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete platform", err.Error())
	}

	return response.Success(c, "Platform deleted successfully", nil)
}

func (h *PlatformHandler) BulkDeletePlatforms(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req struct {
		IDs []uuid.UUID `json:"ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.service.BulkDeletePlatforms(ctx, req.IDs); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete platforms", err.Error())
	}

	return response.Success(c, "Platforms deleted successfully", nil)
}

func (h *PlatformHandler) UploadThumbnail(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "ID is required", nil)
	}
	platformID, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID format", err.Error())
	}

	fileHeader, err := c.FormFile("thumbnail")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Thumbnail file is required", err.Error())
	}

	if fileHeader.Size > 5*1024*1024 { // 5MB limit
		return response.Error(c, fiber.StatusBadRequest, "File size exceeds 5MB limit", nil)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to open file", err.Error())
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	thumbnailURL, err := h.service.UploadThumbnail(ctx, platformID, file, fileHeader.Filename, fileHeader.Size, contentType)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to upload thumbnail", err.Error())
	}

	return response.Success(c, "Thumbnail uploaded successfully", fiber.Map{
		"thumbnail_url": thumbnailURL,
	})
}

func (h *PlatformHandler) GetPlatformByType(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	type_ := c.Params("type")
	if type_ == "" {
		return response.Error(c, fiber.StatusBadRequest, "Type is required", nil)
	}

	platform, err := h.service.GetPlatformByType(ctx, type_)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch platform", err.Error())
	}

	return response.Success(c, "Platform retrieved successfully", platform)
}
