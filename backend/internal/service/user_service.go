package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
	"github.com/user/video-downloader-backend/pkg/utils"
)

type UserService interface {
	FindByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, file io.Reader, filename string, size int64, contentType string) (string, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, req model.UpdatePasswordRequest) error
}

type userService struct {
	repo          repository.UserRepository
	storageClient infrastructure.StorageClient
	cfg           *config.Config
}

func NewUserService(repo repository.UserRepository, storageClient infrastructure.StorageClient, cfg *config.Config) UserService {
	return &userService{
		repo:          repo,
		storageClient: storageClient,
		cfg:           cfg,
	}
}
func (s *userService) FindByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	return s.repo.FindByID(subCtx, userID)
}

func (s *userService) UploadAvatar(ctx context.Context, userID uuid.UUID, file io.Reader, filename string, size int64, contentType string) (string, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	currentUser, err := s.repo.FindByID(subCtx, userID)
	if err == nil && currentUser != nil && currentUser.AvatarURL != nil && *currentUser.AvatarURL != "" {
		oldAvatarURL := *currentUser.AvatarURL

		parsedURL, err := url.Parse(oldAvatarURL)
		if err == nil {
			path := parsedURL.Path
			path = strings.TrimPrefix(path, "/")

			if strings.HasPrefix(path, s.cfg.MinioBucket+"/") {
				objectName := strings.TrimPrefix(path, s.cfg.MinioBucket+"/")

				log.Info().Str("userID", userID.String()).Str("object", objectName).Msg("Deleting old avatar")
				if err := s.storageClient.DeleteFile(subCtx, s.cfg.MinioBucket, objectName); err != nil {
					log.Error().Err(err).Str("object", objectName).Msg("Failed to delete old avatar")
				}
			}
		}
	}

	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("avatars/%s_%d%s", userID.String(), time.Now().Unix(), ext)

	uploadedPath, err := s.storageClient.UploadFile(subCtx, s.cfg.MinioBucket, objectName, file, size, contentType)
	if err != nil {
		return "", fmt.Errorf("storage upload failed: %w", err)
	}

	if err := s.repo.UpdateAvatar(subCtx, userID, uploadedPath); err != nil {
		return "", fmt.Errorf("db update failed: %w", err)
	}

	presignedURL, err := s.storageClient.GetFileURL(subCtx, s.cfg.MinioBucket, uploadedPath, 7*24*time.Hour)
	if err != nil {
		return uploadedPath, nil
	}

	return presignedURL, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	user, err := s.repo.FindByID(subCtx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if req.Email != user.Email {
		existingUser, err := s.repo.FindByEmail(subCtx, req.Email)
		if err != nil {
			return err
		}
		if existingUser != nil {
			return fmt.Errorf("email already in use")
		}
	}

	return s.repo.UpdateProfile(subCtx, userID, req)
}

func (s *userService) UpdatePassword(ctx context.Context, userID uuid.UUID, req model.UpdatePasswordRequest) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	user, err := s.repo.FindByID(subCtx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if user.PasswordHash == nil {
		return fmt.Errorf("user has no password set (oauth account)")
	}

	if !utils.CheckPasswordHash(req.CurrentPassword, *user.PasswordHash) {
		return fmt.Errorf("incorrect current password")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.UpdatePassword(subCtx, userID, hashedPassword)
}
