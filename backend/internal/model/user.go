package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash *string    `json:"-" db:"password_hash"` // Nullable
	FullName     string     `json:"full_name" db:"full_name"`
	AvatarURL    *string    `json:"avatar_url" db:"avatar_url"` // Changed to *string to match DB nullable
	RoleID       *uuid.UUID `json:"role_id" db:"role_id"`       // Changed to *uuid.UUID to match DB nullable
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"` // Soft delete

	// Relations
	Role           *Role           `json:"role,omitempty" db:"-"`
	OAuthProviders []OAuthProvider `json:"oauth_providers,omitempty" db:"-"`
	Downloads      []DownloadTask  `json:"downloads,omitempty" db:"-"`
	Subscriptions  []Subscription  `json:"subscriptions,omitempty" db:"-"`
	Transactions   []Transaction   `json:"transactions,omitempty" db:"-"`
}

type Role struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Permissions map[string]bool `json:"permissions" db:"permissions"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type OAuthProvider struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Provider       string     `json:"provider" db:"provider"`
	ProviderUserID string     `json:"provider_user_id" db:"provider_user_id"`
	AccessToken    *string    `json:"access_token" db:"access_token"`
	RefreshToken   *string    `json:"refresh_token" db:"refresh_token"`
	ExpiryAt       *time.Time `json:"expiry_at" db:"expiry_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`

	// Relations
	User *User `json:"user,omitempty" db:"-"`
}

type EmailAuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterEmailRequest struct {
	FullName string `json:"full_name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type GoogleAuthRequest struct {
	Credential string `json:"credential" validate:"required"` // Changed from IDToken to match handler
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8,eqfield=NewPassword"`
}
