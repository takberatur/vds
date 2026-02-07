package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/user/video-downloader-backend/internal/dto"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/logger"
	"github.com/user/video-downloader-backend/pkg/response"
	"github.com/user/video-downloader-backend/pkg/utils"
)

type SubscriptionHandler struct {
	svc service.SubscriptionService
}

func NewSubscriptionHandler(svc service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

func (h *SubscriptionHandler) UpsertMobile(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}
	appID, _ := c.Locals("app_id").(uuid.UUID)

	var req dto.MobileSubscriptionUpsertRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	platform := req.Platform
	if platform == "" {
		platform = "android"
	}
	status := req.Status
	if status == "" {
		status = "active"
	}

	start := time.UnixMilli(req.StartTimeMs)
	end := time.UnixMilli(req.EndTimeMs)
	if end.Before(start) {
		end = start
	}

	sub := &model.Subscription{
		UserID:                &userID,
		AppID:                 &appID,
		OriginalTransactionID: req.OriginalTransactionID,
		ProductID:             req.ProductID,
		PurchaseToken:         req.PurchaseToken,
		Platform:              platform,
		StartTime:             start,
		EndTime:               end,
		Status:                status,
		AutoRenew:             req.AutoRenew,
	}

	out, err := h.svc.Upsert(ctx, sub)
	if err != nil {
		logger.NotifyTelegram("[sub] upsert failed user=%s product=%s err=%s", userID.String(), req.ProductID, err.Error())
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save subscription", err.Error())
	}

	logger.NotifyTelegram("[sub] upsert ok user=%s product=%s status=%s end=%s", userID.String(), out.ProductID, out.Status, out.EndTime.UTC().Format(time.RFC3339))
	return response.Success(c, "Subscription saved", out)
}

func (h *SubscriptionHandler) GetCurrentMobile(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}
	appID, ok := c.Locals("app_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	sub, err := h.svc.FindCurrentByUserAndApp(ctx, userID, appID, time.Now())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch subscription", err.Error())
	}
	if sub == nil {
		return response.Success(c, "No subscription", nil)
	}

	return response.Success(c, "Subscription", sub)
}

func (h *SubscriptionHandler) FindAll(c *fiber.Ctx) error {
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

	subs, pagination, err := h.svc.FindAll(ctx, params)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch subscriptions", err.Error())
	}

	return response.SuccessWithMeta(c, "Subscriptions",
		subs,
		pagination,
	)
}

func (h *SubscriptionHandler) Delete(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	subID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid subscription ID", err.Error())
	}

	err = h.svc.Delete(ctx, subID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete subscription", err.Error())
	}

	return response.Success(c, "Subscription deleted", nil)
}

func (h *SubscriptionHandler) BulkDelete(c *fiber.Ctx) error {
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
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete subscriptions", err.Error())
	}

	return response.Success(c, "Subscriptions deleted", nil)
}

func (h *SubscriptionHandler) FindByID(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	subID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid subscription ID", err.Error())
	}

	sub, err := h.svc.FindByID(ctx, subID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch subscription", err.Error())
	}
	if sub == nil {
		return response.Success(c, "No subscription", nil)
	}

	return response.Success(c, "Subscription", sub)
}
