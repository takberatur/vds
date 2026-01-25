package handler

import (
	"bufio"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
	"github.com/user/video-downloader-backend/pkg/utils"
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
func (h *AdminHandler) GetCookies(c *fiber.Ctx) error {
	cookiesPath := "cookies.txt"

	file, err := os.Open(cookiesPath)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to read cookies file", err.Error())
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to read cookies file", err.Error())
	}

	return response.Success(c, "Cookies file retrieved successfully", lines)
}
func (h *AdminHandler) UpdateCookies(c *fiber.Ctx) error {
	var req struct {
		Cookies []string `json:"cookies" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	cookiesPath := "cookies.txt"
	file, err := os.OpenFile(cookiesPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to open cookies file", err.Error())
	}
	defer file.Close()

	for _, cookie := range req.Cookies {
		if _, err := file.WriteString(cookie + "\n"); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to write cookies file", err.Error())
		}
	}

	return response.Success(c, "Cookies file updated successfully", nil)
}
