package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
	"github.com/user/video-downloader-backend/pkg/utils"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req struct {
		Credential string `json:"credential"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Basic validation
	if req.Credential == "" {
		return response.Error(c, fiber.StatusBadRequest, "Credential is required", nil)
	}

	user, accessToken, err := h.authService.VerifyGoogleToken(ctx, req.Credential)
	if err != nil {
		// Log the error for debugging purposes since 401 doesn't show details in standard logger
		fmt.Printf("❌ Google Login Failed: %v\n", err)
		return response.Error(c, fiber.StatusUnauthorized, "Authentication failed: "+err.Error(), nil)
	}

	return response.Success(c, "Login successful", fiber.Map{
		"access_token": accessToken,
		"user":         user,
	})
}

func (h *AuthHandler) LoginEmail(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req model.EmailAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	user, accessToken, err := h.authService.LoginEmail(ctx, req.Email, req.Password)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "Authentication failed: "+err.Error(), nil)
	}

	return response.Success(c, "Login successful", fiber.Map{
		"access_token": accessToken,
		"user":         user,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	if err := h.authService.Logout(ctx, userID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Logout failed", err.Error())
	}

	return response.Success(c, "Logout successful", nil)
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, "Email is required", nil)
	}

	if err := h.authService.ForgotPassword(ctx, req.Email); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to process request", err.Error())
	}

	return response.Success(c, "If your email is registered, you will receive a reset link", nil)
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.Token == "" || req.NewPassword == "" {
		return response.Error(c, fiber.StatusBadRequest, "Token and new password are required", nil)
	}

	if err := h.authService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Failed to reset password", err.Error())
	}

	return response.Success(c, "Password has been reset successfully", nil)
}
