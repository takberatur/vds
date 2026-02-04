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

type ApplicationHandler struct {
	svc service.ApplicationService
}

type mobileApplicationResponse struct {
	ID                        uuid.UUID            `json:"id"`
	Name                      string               `json:"name"`
	PackageName               string               `json:"package_name"`
	Version                   *string              `json:"version"`
	Platform                  string               `json:"platform"`
	EnableMonetization        bool                 `json:"enable_monetization"`
	EnableAdmob               bool                 `json:"enable_admob"`
	EnableUnityAd             bool                 `json:"enable_unity_ad"`
	EnableStartApp            bool                 `json:"enable_start_app"`
	EnableInAppPurchase       bool                 `json:"enable_in_app_purchase"`
	AdmobAdUnitID             *string              `json:"admob_ad_unit_id"`
	UnityAdUnitID             *string              `json:"unity_ad_unit_id"`
	StartAppAdUnitID          *string              `json:"start_app_ad_unit_id"`
	AdmobBannerAdUnitID       *string              `json:"admob_banner_ad_unit_id"`
	AdmobInterstitialAdUnitID *string              `json:"admob_interstitial_ad_unit_id"`
	AdmobNativeAdUnitID       *string              `json:"admob_native_ad_unit_id"`
	AdmobRewardedAdUnitID     *string              `json:"admob_rewarded_ad_unit_id"`
	UnityBannerAdUnitID       *string              `json:"unity_banner_ad_unit_id"`
	UnityInterstitialAdUnitID *string              `json:"unity_interstitial_ad_unit_id"`
	UnityNativeAdUnitID       *string              `json:"unity_native_ad_unit_id"`
	UnityRewardedAdUnitID     *string              `json:"unity_rewarded_ad_unit_id"`
	OneSignalID               *string              `json:"one_signal_id"`
	IsActive                  bool                 `json:"is_active"`
	CreatedAt                 time.Time            `json:"created_at"`
	UpdatedAt                 time.Time            `json:"updated_at"`
	InAppProducts             []model.InAppProduct `json:"in_app_products,omitempty"`
}

func NewApplicationHandler(svc service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

func (h *ApplicationHandler) RegisterApp(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req model.RegisterAppRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.Name == "" || req.PackageName == "" {
		return response.Error(c, fiber.StatusBadRequest, "Name and Package Name are required", nil)
	}

	result, err := h.svc.RegisterApp(ctx, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Created(c, "Application registered successfully", result)
}

func (h *ApplicationHandler) GetApplications(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

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

	resp, err := h.svc.FindAll(ctx, params)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch applications", err.Error())
	}

	return response.SuccessWithMeta(c, "Applications retrieved successfully",
		resp.Data,
		resp.Pagination,
	)
}

func (h *ApplicationHandler) UpdateApp(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "Application ID is required", nil)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid application ID", nil)
	}

	var req model.RegisterAppRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	app, err := h.svc.UpdateApp(ctx, id, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, "Application updated successfully", app)
}

func (h *ApplicationHandler) DeleteApp(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "Application ID is required", nil)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid application ID", nil)
	}

	if err := h.svc.DeleteApp(ctx, id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, "Application deleted successfully", nil)
}

func (h *ApplicationHandler) BulkDeleteApps(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req model.BulkDeleteAppsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if len(req.IDs) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "No IDs provided", nil)
	}

	if err := h.svc.BulkDeleteApps(ctx, req.IDs); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to bulk delete applications", err.Error())
	}

	return response.Success(c, "Applications deleted successfully", nil)
}

func (h *ApplicationHandler) FindByID(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
	idStr := c.Params("id")
	if idStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "Application ID is required", nil)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid application ID", nil)
	}

	app, err := h.svc.FindByID(ctx, id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, "Application found successfully", app)
}

func (h *ApplicationHandler) GetCurrent(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
	appID, ok := c.Locals("app_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	app, err := h.svc.FindByID(ctx, appID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch application", err.Error())
	}
	if app == nil {
		return response.Error(c, fiber.StatusNotFound, "Application not found", nil)
	}

	resp := mobileApplicationResponse{
		ID:                        app.ID,
		Name:                      app.Name,
		PackageName:               app.PackageName,
		Version:                   app.Version,
		Platform:                  app.Platform,
		EnableMonetization:        app.EnableMonetization,
		EnableAdmob:               app.EnableAdmob,
		EnableUnityAd:             app.EnableUnityAd,
		EnableStartApp:            app.EnableStartApp,
		EnableInAppPurchase:       app.EnableInAppPurchase,
		AdmobAdUnitID:             app.AdmobAdUnitID,
		UnityAdUnitID:             app.UnityAdUnitID,
		StartAppAdUnitID:          app.StartAppAdUnitID,
		AdmobBannerAdUnitID:       app.AdmobBannerAdUnitID,
		AdmobInterstitialAdUnitID: app.AdmobInterstitialAdUnitID,
		AdmobNativeAdUnitID:       app.AdmobNativeAdUnitID,
		AdmobRewardedAdUnitID:     app.AdmobRewardedAdUnitID,
		UnityBannerAdUnitID:       app.UnityBannerAdUnitID,
		UnityInterstitialAdUnitID: app.UnityInterstitialAdUnitID,
		UnityNativeAdUnitID:       app.UnityNativeAdUnitID,
		UnityRewardedAdUnitID:     app.UnityRewardedAdUnitID,
		IsActive:                  app.IsActive,
		CreatedAt:                 app.CreatedAt,
		UpdatedAt:                 app.UpdatedAt,
		InAppProducts:             app.InAppProducts,
	}

	return response.Success(c, "Application retrieved successfully", resp)
}
