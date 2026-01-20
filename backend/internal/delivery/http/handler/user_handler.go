package handler

import (
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
