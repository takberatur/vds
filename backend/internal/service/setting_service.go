package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
	"github.com/user/video-downloader-backend/pkg/utils"
)

type SettingService interface {
	GetPublicSettings(ctx context.Context) (*model.SettingsResponse, error)
	GetAllSettings(ctx context.Context) ([]model.Setting, error)
	UpdateSetting(ctx context.Context, key, value string) error
	UpdateSettingsBulk(ctx context.Context, settings []model.UpdateSettingsBulkRequest) error
	UploadFile(ctx context.Context, file *multipart.FileHeader, key string) (string, error)
}

type settingService struct {
	repo    repository.SettingRepository
	storage infrastructure.StorageClient
	cfg     *config.Config
}

func NewSettingService(repo repository.SettingRepository, storage infrastructure.StorageClient, cfg *config.Config) SettingService {
	return &settingService{
		repo:    repo,
		storage: storage,
		cfg:     cfg,
	}
}

func (s *settingService) UploadFile(ctx context.Context, file *multipart.FileHeader, key string) (string, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("invalid file type: %s", contentType)
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	bucketName := "video-downloader"
	objectName := fmt.Sprintf("settings/%s-%s", key, file.Filename)

	if oldSetting, err := s.repo.GetByKey(subCtx, key); err == nil && oldSetting.Value != "" {
		oldObject := oldSetting.Value

		parsedURL, err := url.Parse(oldObject)
		if err == nil {
			path := parsedURL.Path
			path = strings.TrimPrefix(path, "/")

			if strings.HasPrefix(path, s.cfg.MinioBucket+"/") {
				objectName := strings.TrimPrefix(path, s.cfg.MinioBucket+"/")

				if err := s.storage.DeleteFile(subCtx, s.cfg.MinioBucket, objectName); err != nil {
					log.Error().Err(err).Str("object", objectName).Msg("Failed to delete old file")
				}
			}
		}
	}

	uploadedPath, err := s.storage.UploadFile(subCtx, bucketName, objectName, src, file.Size, contentType)
	if err != nil {
		return "", err
	}

	if err := s.repo.UpdateByKey(subCtx, key, uploadedPath); err != nil {
		return "", err
	}

	return uploadedPath, nil
}

func (s *settingService) GetPublicSettings(ctx context.Context) (*model.SettingsResponse, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	settings, err := s.repo.GetAll(subCtx)
	if err != nil {
		return nil, err
	}

	response := &model.SettingsResponse{}

	for _, item := range settings {
		switch item.GroupName {
		case "WEBSITE":
			mapWebsiteSetting(&response.WEBSITE, item)
		case "EMAIL":
			mapEmailSetting(&response.EMAIL, item)
		case "SYSTEM":
			mapSystemSetting(&response.SYSTEM, item)
		case "MONETIZE":
			mapMonetizeSetting(&response.MONETIZE, item)
		}
	}

	return response, nil
}

func (s *settingService) GetAllSettings(ctx context.Context) ([]model.Setting, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.GetAll(subCtx)
}

func (s *settingService) UpdateSetting(ctx context.Context, key, value string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()
	return s.repo.UpdateByKey(subCtx, key, value)
}

func (s *settingService) UpdateSettingsBulk(ctx context.Context, settings []model.UpdateSettingsBulkRequest) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	payload := make([]model.Setting, 0, len(settings))
	for _, item := range settings {
		payload = append(payload, model.Setting{
			Key:         item.Key,
			Value:       item.Value,
			Description: item.Description,
			GroupName:   item.GroupName,
		})
	}
	return s.repo.UpdateBulk(subCtx, payload)
}

// Helpers to map dynamic KV to struct fields
func mapWebsiteSetting(target *model.SettingWeb, s model.Setting) {
	if target.SiteCreatedAt.IsZero() {
		target.SiteCreatedAt = s.CreatedAt.In(time.UTC)
	}

	switch s.Key {
	case "site_name":
		target.SiteName = s.Value
	case "site_tagline":
		target.SiteTagline = s.Value
	case "site_description":
		target.SiteDescription = s.Value
	case "site_keywords":
		target.SiteKeywords = s.Value
	case "site_logo":
		target.SiteLogo = s.Value
	case "site_favicon":
		target.SiteFavicon = s.Value
	case "site_email":
		target.SiteEmail = s.Value
	case "site_phone":
		target.SitePhone = s.Value
	case "site_url":
		target.SiteURL = s.Value
	case "site_created_at":
		target.SiteCreatedAt = s.CreatedAt.In(time.UTC)
	}
}

func mapEmailSetting(target *model.SettingEmail, s model.Setting) {
	switch s.Key {
	case "smtp_enabled":
		target.SMTPEnabled = s.Value == "true"
	case "smtp_service":
		target.SMTPService = s.Value
	case "smtp_host":
		target.SMTPHost = s.Value
	case "smtp_port":
		if port, err := strconv.Atoi(s.Value); err == nil {
			target.SMTPPort = port
		}
	case "smtp_user":
		target.SMTPUser = s.Value
	case "smtp_password":
		hash, err := utils.HashPassword(s.Value)
		if err != nil {
			log.Error().Err(err).Str("password", s.Value).Msg("Failed to hash password")
		}
		target.SMTPPassword = hash
	case "from_email":
		target.FromEmail = s.Value
	case "from_name":
		target.FromName = s.Value
	}
}

func mapSystemSetting(target *model.SettingSystem, s model.Setting) {
	switch s.Key {
	case "enable_documentation":
		target.EnableDocumentation = s.Value == "true"
	case "maintenance_mode":
		target.MaintenanceMode = s.Value == "true"
	case "maintenance_message":
		target.MaintenanceMessage = s.Value
	case "source_logo_favicon":
		target.SourceLogoFavicon = s.Value
	case "histats_tracking_code":
		target.HistatsTrackingCode = s.Value
	case "google_analytics_code":
		target.GoogleAnalyticsCode = s.Value
	case "play_store_app_url":
		target.PlayStoreAppURL = s.Value
	case "app_store_app_url":
		target.AppStoreAppURL = s.Value
	}
}

func mapMonetizeSetting(target *model.SettingMonetize, s model.Setting) {
	switch s.Key {
	case "enable_monetize":
		target.EnableMonetize = s.Value == "true"
	case "type_monetize":
		target.TypeMonetize = s.Value
	case "publisher_id":
		target.PublisherID = s.Value
	case "enable_popup_ad":
		target.EnablePopupAd = s.Value == "true"
	case "enable_socialbar_ad":
		target.EnableSocialbarAd = s.Value == "true"
	case "auto_ad_code":
		target.AutoAdCode = s.Value
	case "popup_ad_code":
		target.PopupAdCode = s.Value
	case "socialbar_ad_code":
		target.SocialbarAdCode = s.Value
	case "banner_rectangle_ad_code":
		target.BannerRectangleAdCode = s.Value
	case "banner_horizontal_ad_code":
		target.BannerHorizontalAdCode = s.Value
	case "banner_vertical_ad_code":
		target.BannerVerticalAdCode = s.Value
	case "native_ad_code":
		target.NativeAdCode = s.Value
	case "direct_link_ad_code":
		target.DirectLinkAdCode = s.Value
	}
}
