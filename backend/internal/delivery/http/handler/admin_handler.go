package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}
func (h *AdminHandler) GetDashboardData(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	params := model.QueryParamsRequest{}

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

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	params.Limit = limit

	resp, err := h.adminService.GetDashboardData(ctx, params)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to get dashboard data", err.Error())
	}

	return response.SuccessWithMeta(c, "Dashboard data retrieved successfully", resp.Data, resp.Pagination)
}
