package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/model"
)

type TokenService interface {
	GenerateAccessToken(user *model.User) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}

type tokenService struct {
	cfg *config.Config
}

func NewTokenService(cfg *config.Config) TokenService {
	return &tokenService{cfg: cfg}
}

func (s *tokenService) GenerateAccessToken(user *model.User) (string, error) {
	expiryHour, _ := strconv.Atoi(s.cfg.JWTExpiryHour)
	expiryTime := time.Now().Add(time.Duration(expiryHour) * time.Hour)

	var roleName string
	if user.Role != nil {
		roleName = user.Role.Name
	}

	claims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"email":     user.Email,
		"role":      user.RoleID,
		"role_name": roleName,
		"exp":       expiryTime.Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// IMPORTANT: Ensure the secret is NOT empty
	if s.cfg.JWTSecret == "" {
		return "", fmt.Errorf("JWT secret is not configured")
	}
	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (s *tokenService) GenerateRefreshToken(user *model.User) (string, error) {
	expiryHour, _ := strconv.Atoi(s.cfg.JWTExpiryHour)
	expiryTime := time.Now().Add(time.Duration(expiryHour) * time.Hour)

	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": expiryTime.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (s *tokenService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if s.cfg.JWTSecret == "" {
			return nil, fmt.Errorf("JWT secret is not configured in validator")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
}
