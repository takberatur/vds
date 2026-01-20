package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/dto"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
)

type AdminRepository interface {
	BaseRepository
	GetDashboardData(ctx context.Context, params model.QueryParamsRequest) (*dto.DashboardResponse, error)
}

type adminRepository struct {
	*baseRepository
}

func NewAdminRepository(db *pgxpool.Pool) AdminRepository {
	return &adminRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}

func (r *adminRepository) GetDashboardData(ctx context.Context, params model.QueryParamsRequest) (*dto.DashboardResponse, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	var stats dto.DashboardStats

	if err := r.db.QueryRow(subCtx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers); err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	if err := r.db.QueryRow(subCtx, "SELECT COUNT(*) FROM applications").Scan(&stats.TotalApps); err != nil {
		return nil, fmt.Errorf("failed to count apps: %w", err)
	}
	if err := r.db.QueryRow(subCtx, "SELECT COUNT(*) FROM platforms").Scan(&stats.TotalPlatforms); err != nil {
		return nil, fmt.Errorf("failed to count platforms: %w", err)
	}
	if err := r.db.QueryRow(subCtx, "SELECT COUNT(*) FROM downloads").Scan(&stats.TotalDownloads); err != nil {
		return nil, fmt.Errorf("failed to count downloads: %w", err)
	}
	if err := r.db.QueryRow(subCtx, "SELECT COUNT(*) FROM subscriptions").Scan(&stats.TotalSubscriptions); err != nil {
		return nil, fmt.Errorf("failed to count subscriptions: %w", err)
	}
	if err := r.db.QueryRow(subCtx, "SELECT COUNT(*) FROM transactions").Scan(&stats.TotalTransactions); err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	dateFrom := params.DateFrom
	dateTo := params.DateTo

	if dateFrom.IsZero() {
		dateFrom = time.Now().AddDate(0, 0, -30)
	}
	if dateTo.IsZero() {
		dateTo = time.Now()
	}

	analyticsQuery := `
		SELECT id, date, total_downloads, total_users, active_users, total_revenue, updated_at
		FROM analytics_daily
		WHERE date >= $1 AND date <= $2
		ORDER BY date ASC
	`
	rowsA, err := r.db.Query(subCtx, analyticsQuery, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to query analytics: %w", err)
	}
	defer rowsA.Close()

	var analytics []model.AnalyticsDaily
	for rowsA.Next() {
		var a model.AnalyticsDaily
		if err := rowsA.Scan(&a.ID, &a.Date, &a.TotalDownloads, &a.TotalUsers, &a.ActiveUsers, &a.TotalRevenue, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan analytics: %w", err)
		}
		analytics = append(analytics, a)
	}
	if analytics == nil {
		analytics = []model.AnalyticsDaily{}
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	downloadsQuery := `
		SELECT id, user_id, original_url, platform_id, status, file_path, format, thumbnail_url, title, file_size, duration, created_at
		FROM downloads
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
		LIMIT $3
	`
	rowsD, err := r.db.Query(subCtx, downloadsQuery, dateFrom, dateTo, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent downloads: %w", err)
	}
	defer rowsD.Close()

	var recentDownloads []model.DownloadTask
	for rowsD.Next() {
		var d model.DownloadTask
		if err := rowsD.Scan(
			&d.ID, &d.UserID, &d.OriginalURL, &d.PlatformID, &d.Status, &d.FilePath, &d.Format, &d.ThumbnailURL,
			&d.Title, &d.FileSize, &d.Duration, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan download task: %w", err)
		}
		recentDownloads = append(recentDownloads, d)
	}
	if recentDownloads == nil {
		recentDownloads = []model.DownloadTask{}
	}

	response := &dto.DashboardResponse{
		Data: dto.DashboardData{
			Stats:           stats,
			Analytics:       analytics,
			RecentDownloads: recentDownloads,
		},
		Pagination: dto.Pagination{
			Limit: limit,
		},
	}

	return response, nil
}
