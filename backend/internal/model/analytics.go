package model

import (
	"time"

	"github.com/google/uuid"
)

type AnalyticsDaily struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Date           time.Time `json:"date" db:"date"`
	TotalDownloads int       `json:"total_downloads" db:"total_downloads"`
	TotalUsers     int       `json:"total_users" db:"total_users"`
	ActiveUsers    int       `json:"active_users" db:"active_users"`
	TotalRevenue   float64   `json:"total_revenue" db:"total_revenue"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
