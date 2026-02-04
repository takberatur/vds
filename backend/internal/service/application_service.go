package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
)

type ApplicationService interface {
	RegisterApp(ctx context.Context, req model.RegisterAppRequest) (*model.RegisterAppResponse, error)
	FindAll(ctx context.Context, params model.QueryParamsRequest) (*model.ApplicationsResponse, error)
	UpdateApp(ctx context.Context, id uuid.UUID, req model.RegisterAppRequest) (*model.Application, error)
	DeleteApp(ctx context.Context, id uuid.UUID) error
	BulkDeleteApps(ctx context.Context, ids []uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Application, error)
}

type applicationService struct {
	repo repository.ApplicationRepository
}

func NewApplicationService(repo repository.ApplicationRepository) ApplicationService {
	return &applicationService{repo: repo}
}

func (s *applicationService) RegisterApp(ctx context.Context, req model.RegisterAppRequest) (*model.RegisterAppResponse, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	existingApp, err := s.repo.FindByPackageName(subCtx, req.PackageName)
	if err != nil {
		return nil, err
	}
	if existingApp != nil {
		return nil, errors.New("application with this package name already exists")
	}

	apiKey, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}

	secretKey, err := generateRandomString(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}

	if req.Platform == "" {
		req.Platform = "android"
	}

	var version *string
	if req.Version != "" {
		version = &req.Version
	}

	app := &model.Application{
		Name:                      req.Name,
		PackageName:               req.PackageName,
		APIKey:                    apiKey,
		SecretKey:                 secretKey,
		Version:                   version,
		Platform:                  req.Platform,
		EnableMonetization:        req.EnableMonetization,
		EnableAdmob:               req.EnableAdmob,
		EnableUnityAd:             req.EnableUnityAd,
		EnableStartApp:            req.EnableStartApp,
		EnableInAppPurchase:       req.EnableInAppPurchase,
		AdmobAdUnitID:             req.AdmobAdUnitID,
		UnityAdUnitID:             req.UnityAdUnitID,
		StartAppAdUnitID:          req.StartAppAdUnitID,
		AdmobBannerAdUnitID:       req.AdmobBannerAdUnitID,
		AdmobInterstitialAdUnitID: req.AdmobInterstitialAdUnitID,
		AdmobNativeAdUnitID:       req.AdmobNativeAdUnitID,
		AdmobRewardedAdUnitID:     req.AdmobRewardedAdUnitID,
		UnityBannerAdUnitID:       req.UnityBannerAdUnitID,
		UnityInterstitialAdUnitID: req.UnityInterstitialAdUnitID,
		UnityNativeAdUnitID:       req.UnityNativeAdUnitID,
		UnityRewardedAdUnitID:     req.UnityRewardedAdUnitID,
		OneSignalID:               req.OneSignalID,
		IsActive:                  req.IsActive,
	}

	if err := s.repo.Create(subCtx, app); err != nil {
		return nil, err
	}

	return &model.RegisterAppResponse{
		Name:        app.Name,
		PackageName: app.PackageName,
		APIKey:      app.APIKey,
		SecretKey:   app.SecretKey,
	}, nil
}
func (s *applicationService) FindAll(ctx context.Context, params model.QueryParamsRequest) (*model.ApplicationsResponse, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	apps, pagination, err := s.repo.FindAll(subCtx, params)
	if err != nil {
		return nil, err
	}
	return &model.ApplicationsResponse{
		Data:       apps,
		Pagination: pagination,
	}, nil
}

func (s *applicationService) UpdateApp(ctx context.Context, id uuid.UUID, req model.RegisterAppRequest) (*model.Application, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	app, err := s.repo.FindByID(subCtx, id)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	app.Name = req.Name
	app.PackageName = req.PackageName

	if req.Version != "" {
		app.Version = &req.Version
	} else {
		app.Version = nil
	}

	app.Platform = req.Platform
	app.EnableMonetization = req.EnableMonetization
	app.EnableAdmob = req.EnableAdmob
	app.EnableUnityAd = req.EnableUnityAd
	app.EnableStartApp = req.EnableStartApp
	app.EnableInAppPurchase = req.EnableInAppPurchase
	app.AdmobAdUnitID = req.AdmobAdUnitID
	app.UnityAdUnitID = req.UnityAdUnitID
	app.StartAppAdUnitID = req.StartAppAdUnitID
	app.AdmobBannerAdUnitID = req.AdmobBannerAdUnitID
	app.AdmobInterstitialAdUnitID = req.AdmobInterstitialAdUnitID
	app.AdmobNativeAdUnitID = req.AdmobNativeAdUnitID
	app.AdmobRewardedAdUnitID = req.AdmobRewardedAdUnitID
	app.UnityBannerAdUnitID = req.UnityBannerAdUnitID
	app.UnityInterstitialAdUnitID = req.UnityInterstitialAdUnitID
	app.UnityNativeAdUnitID = req.UnityNativeAdUnitID
	app.UnityRewardedAdUnitID = req.UnityRewardedAdUnitID
	app.OneSignalID = req.OneSignalID
	app.IsActive = req.IsActive

	if err := s.repo.Update(subCtx, app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *applicationService) DeleteApp(ctx context.Context, id uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	app, err := s.repo.FindByID(subCtx, id)
	if err != nil {
		return err
	}
	if app == nil {
		return errors.New("application not found")
	}
	return s.repo.Delete(subCtx, id)
}

func (s *applicationService) BulkDeleteApps(ctx context.Context, ids []uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.BulkDelete(subCtx, ids)
}

func (s *applicationService) FindByID(ctx context.Context, id uuid.UUID) (*model.Application, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.FindByID(subCtx, id)
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
