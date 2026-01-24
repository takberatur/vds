package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
)

type CentrifugoHandler struct {
	tokenService service.TokenService
}

func NewCentrifugoHandler(tokenService service.TokenService) *CentrifugoHandler {
	return &CentrifugoHandler{
		tokenService: tokenService,
	}
}

// GetToken generates a connection token for Centrifugo
func (h *CentrifugoHandler) GetToken(c *fiber.Ctx) error {
	// Check if user is authenticated (optional)
	userID := ""
	if uid := c.Locals("user_id"); uid != nil {
		// If authenticated, use user ID
		// Convert uid to string properly based on type
		// Assuming uid is string or int64 from middleware
		userID = fmt.Sprintf("%v", uid)
	}

	// Generate token
	token, err := h.tokenService.GenerateCentrifugoToken(userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to generate connection token", err.Error())
	}

	return response.Success(c, "Token generated successfully", fiber.Map{
		"token": token,
	})
}
