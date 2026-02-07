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
	"github.com/user/video-downloader-backend/pkg/utils"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	user, err := h.svc.FindByID(ctx, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to find user", err.Error())
	}

	return response.Success(c, "User retrieved successfully", user)
}

func (h *UserHandler) UploadAvatar(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Avatar file is required", err.Error())
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
	avatarURL, err := h.svc.UploadAvatar(ctx, userID, file, fileHeader.Filename, fileHeader.Size, contentType)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to upload avatar", err.Error())
	}

	return response.Success(c, "Avatar uploaded successfully", fiber.Map{
		"avatar_url": avatarURL,
	})
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req model.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	if err := h.svc.UpdateProfile(ctx, userID, req); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update profile", err.Error())
	}

	return response.Success(c, "Profile updated successfully", nil)
}

func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req model.UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	if err := h.svc.UpdatePassword(ctx, userID, req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Failed to update password", err.Error())
	}

	return response.Success(c, "Password updated successfully", nil)
}

func (h *UserHandler) FindAll(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	params := model.QueryParamsRequest{
		Search:   c.Query("search"),
		SortBy:   c.Query("sort_by", "created_at"),
		OrderBy:  c.Query("order_by", "desc"),
		Page:     page,
		Limit:    limit,
		Status:   c.Query("status"),
		IsActive: c.Query("is_active") == "true",
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

	subs, pagination, err := h.svc.FindAll(ctx, params)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch users", err.Error())
	}

	return response.SuccessWithMeta(c, "Users",
		subs,
		pagination,
	)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "User ID is required", nil)
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
	}

	err = h.svc.Delete(ctx, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete user", err.Error())
	}

	return response.Success(c, "User deleted", nil)
}

func (h *UserHandler) BulkDelete(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req struct {
		IDs []string `json:"ids" validate:"required,dive,uuid4"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	subIDs := make([]uuid.UUID, 0, len(req.IDs))
	for _, id := range req.IDs {
		if subID, err := uuid.Parse(id); err == nil {
			subIDs = append(subIDs, subID)
		}
	}

	err := h.svc.BulkDelete(ctx, subIDs)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete users", err.Error())
	}

	return response.Success(c, "Users deleted", nil)
}

func (h *UserHandler) FindByID(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "User ID is required", nil)
	}

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
	}

	user, err := h.svc.FindByID(ctx, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch user", err.Error())
	}
	if user == nil {
		return response.Success(c, "No user", nil)
	}

	return response.Success(c, "User", user)
}
