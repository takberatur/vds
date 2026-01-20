package dto

import "github.com/user/video-downloader-backend/internal/model"

type DashboardResponse struct {
	Data       DashboardData `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

type DashboardData struct {
	Stats           DashboardStats         `json:"stats"`
	Analytics       []model.AnalyticsDaily `json:"analytics"`
	RecentDownloads []model.DownloadTask   `json:"recent_downloads"`
}

type DashboardStats struct {
	TotalUsers         int64 `json:"total_users"`
	TotalApps          int64 `json:"total_apps"`
	TotalPlatforms     int64 `json:"total_platforms"`
	TotalDownloads     int64 `json:"total_downloads"`
	TotalSubscriptions int64 `json:"total_subscriptions"`
	TotalTransactions  int64 `json:"total_transactions"`
}

type PlatformsResponse struct {
	Data       []model.Platform `json:"data"`
	Pagination Pagination       `json:"pagination"`
}
