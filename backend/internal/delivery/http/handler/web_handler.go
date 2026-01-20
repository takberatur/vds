package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/video-downloader-backend/internal/dto"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/service"
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
