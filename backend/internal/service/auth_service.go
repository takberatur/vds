package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/user/video-downloader-backend/internal/delivery/helpers"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
	"github.com/user/video-downloader-backend/pkg/utils"
	"google.golang.org/api/idtoken"
)

type AuthService interface {
	VerifyGoogleToken(ctx context.Context, idToken string) (*model.User, string, error)
	LoginEmail(ctx context.Context, email, password string) (*model.User, string, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type authService struct {
	userRepo     repository.UserRepository
	mailHelper   helpers.MailHelper
	tokenService TokenService
	redisClient  *redis.Client
}

func NewAuthService(userRepo repository.UserRepository, mailHelper helpers.MailHelper, tokenService TokenService, redisClient *redis.Client) AuthService {
	return &authService{
		userRepo:     userRepo,
		mailHelper:   mailHelper,
		tokenService: tokenService,
		redisClient:  redisClient,
	}
}

func (s *authService) VerifyGoogleToken(ctx context.Context, idToken string) (*model.User, string, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	payload, err := idtoken.Validate(subCtx, idToken, "")
	if err != nil {
		return nil, "", fmt.Errorf("invalid google token: %w", err)
	}

	email := payload.Claims["email"].(string)
	name := payload.Claims["name"].(string)
	picture := payload.Claims["picture"].(string)

	user, err := s.userRepo.FindByEmail(subCtx, email)
	if err != nil {
		return nil, "", err
	}

	if user == nil {
		avatarURL := picture

		user = &model.User{
			Email:     email,
			FullName:  name,
			AvatarURL: &avatarURL,
			RoleID:    nil,
			IsActive:  true,
		}
		if err := s.userRepo.Create(subCtx, user, ""); err != nil {
			return nil, "", err
		}

		fullUser, err := s.userRepo.FindByID(subCtx, user.ID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to fetch user details after creation: %w", err)
		}
		user = fullUser
	}

	_ = s.userRepo.UpdateLastLogin(subCtx, user.ID)

	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, accessToken, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *authService) LoginEmail(ctx context.Context, email, password string) (*model.User, string, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	user, err := s.userRepo.FindByEmail(subCtx, email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if user.PasswordHash == nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, *user.PasswordHash) {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	_ = s.userRepo.UpdateLastLogin(subCtx, user.ID)

	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, accessToken, nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	user, err := s.userRepo.FindByEmail(subCtx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	resetToken, err := utils.GenerateRandomString(32)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("reset_token:%s", resetToken)
	if err := s.redisClient.Set(subCtx, key, user.ID, time.Hour).Err(); err != nil {
		return err
	}

	if err := s.mailHelper.SendResetPasswordEmail(subCtx, user.Email, resetToken); err != nil {
		return err
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	key := fmt.Sprintf("reset_token:%s", token)
	userIDStr, err := s.redisClient.Get(subCtx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("invalid or expired token")
	}
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user id in token: %w", err)
	}

	user, err := s.userRepo.FindByID(subCtx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(subCtx, userID, hashedPassword); err != nil {
		return err
	}

	s.redisClient.Del(subCtx, key)

	return nil
}
