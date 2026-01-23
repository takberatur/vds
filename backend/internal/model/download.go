package model

import (
	"time"

	"github.com/google/uuid"
)

type DownloadFormat struct {
	URL      string   `json:"url"`
	Filesize *int64   `json:"filesize,omitempty"`
	FormatID string   `json:"format_id,omitempty"`
	Acodec   string   `json:"acodec,omitempty"`
	Vcodec   string   `json:"vcodec,omitempty"`
	Ext      string   `json:"ext,omitempty"`
	Height   *int     `json:"height,omitempty"`
	Width    *int     `json:"width,omitempty"`
	Tbr      *float64 `json:"tbr,omitempty"`
}

type DownloadTask struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	UserID        *uuid.UUID       `json:"user_id,omitempty" db:"user_id"`
	AppID         *uuid.UUID       `json:"app_id,omitempty" db:"app_id"`
	PlatformID    uuid.UUID        `json:"platform_id,omitempty" db:"platform_id"`
	PlatformType  string           `json:"platform_type,omitempty" db:"platform_type"`
	OriginalURL   string           `json:"original_url" db:"original_url"`
	FilePath      *string          `json:"file_path" db:"file_path"`
	ThumbnailURL  *string          `json:"thumbnail_url" db:"thumbnail_url"`
	Title         *string          `json:"title" db:"title"`
	Duration      *int             `json:"duration" db:"duration"`
	FileSize      *int64           `json:"file_size" db:"file_size"`
	EncryptedData *[]byte          `json:"encrypted_data" db:"encrypted_data"`
	Format        *string          `json:"format" db:"format"`
	Status        string           `json:"status" db:"status"`
	ErrorMessage  *string          `json:"error_message" db:"error_message"`
	IPAddress     *string          `json:"ip_address" db:"ip_address"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	Formats       []DownloadFormat `json:"formats,omitempty" db:"-"`

	User          *User          `json:"user,omitempty" db:"-"`
	Application   *Application   `json:"application,omitempty" db:"-"`
	Platform      *Platform      `json:"platform,omitempty" db:"-"`
	DownloadFiles []DownloadFile `json:"download_files,omitempty" db:"-"`
}

type DownloadFile struct {
	ID            uuid.UUID `json:"id" db:"id"`
	DownloadID    uuid.UUID `json:"download_id" db:"download_id"`
	URL           string    `json:"url" db:"url"`
	FormatID      *string   `json:"format_id,omitempty" db:"format_id"`
	Resolution    *string   `json:"resolution,omitempty" db:"resolution"`
	Extension     *string   `json:"extension,omitempty" db:"extension"`
	FileSize      *int64    `json:"file_size,omitempty" db:"file_size"`
	EncryptedData *[]byte   `json:"encrypted_data,omitempty" db:"encrypted_data"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`

	DownloadTask *DownloadTask `json:"download_task,omitempty" db:"-"`
}

type Platform struct {
	ID           uuid.UUID      `json:"id" db:"id"`
	Name         string         `json:"name" db:"name"`
	Slug         string         `json:"slug" db:"slug"`
	Type         string         `json:"type" db:"type"`
	ThumbnailURL string         `json:"thumbnail_url" db:"thumbnail_url"`
	URLPattern   *string        `json:"url_pattern" db:"url_pattern"`
	IsActive     bool           `json:"is_active" db:"is_active"`
	IsPremium    bool           `json:"is_premium" db:"is_premium"`
	Config       map[string]any `json:"config" db:"config"` // JSONB
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
}

type DownloadRequest struct {
	URL        string  `json:"url" validate:"required,url"`
	Type       string  `json:"type" validate:"required,oneof=youtube facebook twitter tiktok instagram rumble vimeo dailymotion any-video-downloader youtube-to-mp3 snackvideo"`
	UserID     *string `json:"user_id,omitempty" validate:"omitempty"`
	PlatformID *string `json:"platform_id,omitempty" validate:"omitempty"`
	AppID      *string `json:"app_id,omitempty" validate:"omitempty"`
}

type DownloadPayload struct {
	ID           uuid.UUID        `json:"id,omitempty"`
	Status       string           `json:"status,omitempty"`
	Progress     int              `json:"progress,omitempty"`
	Title        string           `json:"title,omitempty"`
	ThumbnailURL string           `json:"thumbnail_url,omitempty"`
	Type         string           `json:"type,omitempty"`
	CreatedAt    time.Time        `json:"created_at,omitempty"`
	FilePath     *string          `json:"file_path,omitempty"`
	Formats      []DownloadFormat `json:"formats,omitempty"`
}

type DownloadEvent struct {
	Type      string           `json:"type"`
	TaskID    uuid.UUID        `json:"task_id"`
	UserID    *uuid.UUID       `json:"user_id,omitempty"`
	Status    string           `json:"status"`
	Progress  *int             `json:"progress,omitempty"`
	Message   string           `json:"message,omitempty"`
	Error     string           `json:"error,omitempty"`
	Payload   *DownloadPayload `json:"payload,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}
