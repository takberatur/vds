package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
)

type AnalyticRepository interface {
	BaseRepository
	Create(ctx context.Context, analytic *model.AnalyticsDaily) error
}

type analyticRepository struct {
	*baseRepository
}

func NewAnalyticRepository(db *pgxpool.Pool) AnalyticRepository {
	return &analyticRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}

func (r *analyticRepository) Create(ctx context.Context, analytic *model.AnalyticsDaily) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		INSERT INTO analytics_daily (date, total_downloads, total_users, active_users, total_revenue, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(subCtx, query, analytic.Date, analytic.TotalDownloads, analytic.TotalUsers, analytic.ActiveUsers, analytic.TotalRevenue, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create analytic: %w", err)
	}
	return nil
}
