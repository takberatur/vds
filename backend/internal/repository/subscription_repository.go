package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
)

type SubscriptionRepository interface {
	BaseRepository
	Upsert(ctx context.Context, sub *model.Subscription) (*model.Subscription, error)
	FindCurrentByUserAndApp(ctx context.Context, userID uuid.UUID, appID uuid.UUID, now time.Time) (*model.Subscription, error)
	FindByID(ctx context.Context, subID uuid.UUID) (*model.Subscription, error)
	FindAll(ctx context.Context, params model.QueryParamsRequest) ([]model.Subscription, model.Pagination, error)
	Delete(ctx context.Context, subID uuid.UUID) error
	BulkDelete(ctx context.Context, subIDs []uuid.UUID) error
}

type subscriptionRepository struct {
	*baseRepository
}

func NewSubscriptionRepository(db *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}

func (r *subscriptionRepository) Upsert(ctx context.Context, sub *model.Subscription) (*model.Subscription, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	now := time.Now()
	query := `
		INSERT INTO subscriptions (
			user_id,
			app_id,
			original_transaction_id,
			product_id,
			purchase_token,
			platform,
			start_time,
			end_time,
			status,
			auto_renew,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (original_transaction_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			app_id = EXCLUDED.app_id,
			product_id = EXCLUDED.product_id,
			purchase_token = EXCLUDED.purchase_token,
			platform = EXCLUDED.platform,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			status = EXCLUDED.status,
			auto_renew = EXCLUDED.auto_renew,
			updated_at = EXCLUDED.updated_at
		RETURNING *
	`

	var out model.Subscription
	err := pgxscan.Get(
		subCtx,
		r.db,
		&out,
		query,
		sub.UserID,
		sub.AppID,
		sub.OriginalTransactionID,
		sub.ProductID,
		sub.PurchaseToken,
		sub.Platform,
		sub.StartTime,
		sub.EndTime,
		sub.Status,
		sub.AutoRenew,
		now,
		now,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *subscriptionRepository) FindCurrentByUserAndApp(ctx context.Context, userID uuid.UUID, appID uuid.UUID, now time.Time) (*model.Subscription, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		SELECT *
		FROM subscriptions
		WHERE user_id = $1
		  AND app_id = $2
		ORDER BY end_time DESC
		LIMIT 1
	`
	var out model.Subscription
	err := pgxscan.Get(subCtx, r.db, &out, query, userID, appID)
	if err != nil {
		return nil, err
	}

	if out.EndTime.Before(now) && out.Status == "active" {
		out.Status = "expired"
	}
	return &out, nil
}

func (r *subscriptionRepository) FindAll(ctx context.Context, params model.QueryParamsRequest) ([]model.Subscription, model.Pagination, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	qb := NewQueryBuilder(`SELECT id, user_id, app_id, original_transaction_id, product_id, purchase_token, platform, start_time, end_time, status, auto_renew, created_at, updated_at FROM subscriptions`)

	if params.Search != "" {
		qb.Where("(app_id ILIKE $? OR product_id ILIKE $? OR platform ILIKE $?)",
			"%"+params.Search+"%",
			"%"+params.Search+"%",
			"%"+params.Search+"%",
		)
	}

	if params.Status != "" {
		qb.Where("status = $?", params.Status)
	}

	if !params.DateFrom.IsZero() && !params.DateTo.IsZero() {
		qb.Where("created_at BETWEEN $? AND $?", params.DateFrom, params.DateTo)
	}

	if params.SortBy != "" {
		qb.OrderByField(params.SortBy, params.OrderBy)
	} else {
		qb.OrderByField("created_at", "DESC")
	}

	countQuery, countArgs := qb.Clone().ChangeBase("SELECT COUNT(*) FROM subscriptions").WithoutPagination().Build()

	var totalItems int64
	err := pgxscan.Get(subCtx, r.db, &totalItems, countQuery, countArgs...)
	if err != nil {
		return nil, model.Pagination{}, err
	}

	offset := (params.Page - 1) * params.Limit
	qb.WithLimit(params.Limit).WithOffset(offset)

	query, args := qb.Build()
	rows, err := r.db.Query(subCtx, query, args...)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var out []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		if err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.AppID,
			&sub.OriginalTransactionID,
			&sub.ProductID,
			&sub.PurchaseToken,
			&sub.Platform,
			&sub.StartTime,
			&sub.EndTime,
			&sub.Status,
			&sub.AutoRenew,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		); err != nil {
			return nil, model.Pagination{}, fmt.Errorf("failed to scan subscription: %w", err)
		}
		out = append(out, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, model.Pagination{}, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	totalPages := 0
	if params.Limit > 0 {
		totalPages = int((totalItems + int64(params.Limit) - 1) / int64(params.Limit))
	}

	pagination := model.Pagination{
		CurrentPage: params.Page,
		Limit:       params.Limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNext:     params.Page < totalPages,
		HasPrev:     params.Page > 1,
	}
	return out, pagination, nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, subID uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		DELETE FROM subscriptions
		WHERE id = $1
	`
	_, err := r.db.Exec(subCtx, query, subID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

func (r *subscriptionRepository) BulkDelete(ctx context.Context, subIDs []uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		DELETE FROM subscriptions
		WHERE id = ANY($1)
	`
	_, err := r.db.Exec(subCtx, query, subIDs)
	if err != nil {
		return fmt.Errorf("failed to delete subscriptions: %w", err)
	}
	return nil
}

func (r *subscriptionRepository) FindByID(ctx context.Context, subID uuid.UUID) (*model.Subscription, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		SELECT * FROM subscriptions WHERE id = $1
	`
	var out model.Subscription
	err := pgxscan.Get(subCtx, r.db, &out, query, subID)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription by id: %w", err)
	}
	return &out, nil
}
