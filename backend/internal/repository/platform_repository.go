package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
)

type PlatformRepository interface {
	BaseRepository
	GetAll(ctx context.Context) ([]model.Platform, error)
	FindAll(ctx context.Context, params model.QueryParamsRequest) ([]model.Platform, model.Pagination, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Platform, error)
	FindBySlug(ctx context.Context, slug string) (*model.Platform, error)
	FindByType(ctx context.Context, type_ string) (*model.Platform, error)
	Create(ctx context.Context, platform *model.Platform) error
	Update(ctx context.Context, platform *model.Platform) error
	UpdateThumbnail(ctx context.Context, thumbnail string, platformID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
}

type platformRepository struct {
	*baseRepository
}

func NewPlatformRepository(db *pgxpool.Pool) PlatformRepository {
	return &platformRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}

func (r *platformRepository) GetAll(ctx context.Context) ([]model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `SELECT * FROM platforms`

	var platforms []model.Platform
	err := pgxscan.Select(subCtx, r.db, &platforms, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query platforms: %w", err)
	}

	return platforms, nil
}

func (r *platformRepository) FindAll(ctx context.Context, params model.QueryParamsRequest) ([]model.Platform, model.Pagination, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	qb := NewQueryBuilder(`
		SELECT id, name, slug, type, thumbnail_url, url_pattern, is_active, is_premium, config, created_at
		FROM platforms
	`)

	// 1. Filtering
	if params.Search != "" {
		qb.Where("(name ILIKE $? OR slug ILIKE $? OR type ILIKE $?)", "%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	if params.Type != "" {
		qb.Where("type = $?", params.Type)
	}

	if params.Status != "" {
		switch params.Status {
		case "active", "true":
			qb.Where("is_active = true")
		case "inactive", "false":
			qb.Where("is_active = false")
		}
	}

	if !params.DateFrom.IsZero() && !params.DateTo.IsZero() {
		qb.Where("created_at BETWEEN $? AND $?", params.DateFrom, params.DateTo)
	}

	// 2. Sorting
	if params.SortBy != "" {
		qb.OrderByField(params.SortBy, params.OrderBy)
	} else {
		qb.OrderByField("created_at", "DESC")
	}

	// 3. Count Total (before limit/offset)
	countQuery, countArgs := qb.Clone().ChangeBase("SELECT COUNT(*) FROM platforms").WithoutPagination().Build()

	var totalItems int64
	err := r.db.QueryRow(subCtx, countQuery, countArgs...).Scan(&totalItems)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("failed to count platforms: %w", err)
	}

	// 4. Pagination
	offset := (params.Page - 1) * params.Limit
	qb.WithLimit(params.Limit).WithOffset(offset)

	// 5. Execute Query
	query, args := qb.Build()
	rows, err := r.db.Query(subCtx, query, args...)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("failed to query platforms: %w", err)
	}
	defer rows.Close()

	var platforms []model.Platform
	for rows.Next() {
		var p model.Platform
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Type, &p.ThumbnailURL, &p.URLPattern,
			&p.IsActive, &p.IsPremium, &p.Config, &p.CreatedAt,
		); err != nil {
			return nil, model.Pagination{}, err
		}
		platforms = append(platforms, p)
	}

	// 6. Build Pagination Response
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

	return platforms, pagination, nil
}

func (r *platformRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()
	query := `
		SELECT id, name, slug, type, thumbnail_url, url_pattern, is_active, is_premium, config, created_at
		FROM platforms WHERE id = $1
	`
	var p model.Platform
	err := r.db.QueryRow(subCtx, query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Type, &p.ThumbnailURL, &p.URLPattern,
		&p.IsActive, &p.IsPremium, &p.Config, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *platformRepository) FindBySlug(ctx context.Context, slug string) (*model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		SELECT id, name, slug, type, thumbnail_url, url_pattern, is_active, is_premium, config, created_at
		FROM platforms WHERE slug = $1
	`
	var p model.Platform
	err := r.db.QueryRow(subCtx, query, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Type, &p.ThumbnailURL, &p.URLPattern,
		&p.IsActive, &p.IsPremium, &p.Config, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *platformRepository) FindByType(ctx context.Context, type_ string) (*model.Platform, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		SELECT id, name, slug, type, thumbnail_url, url_pattern, is_active, is_premium, config, created_at
		FROM platforms WHERE type = $1
	`
	var p model.Platform
	err := r.db.QueryRow(subCtx, query, type_).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Type, &p.ThumbnailURL, &p.URLPattern,
		&p.IsActive, &p.IsPremium, &p.Config, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *platformRepository) Create(ctx context.Context, platform *model.Platform) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		INSERT INTO platforms (name, slug, type, thumbnail_url, url_pattern, is_active, is_premium, config, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	now := time.Now()
	err := r.db.QueryRow(subCtx, query,
		platform.Name, platform.Slug, platform.Type, platform.ThumbnailURL, platform.URLPattern,
		platform.IsActive, platform.IsPremium, platform.Config, now,
	).Scan(&platform.ID, &platform.CreatedAt)
	return err
}

func (r *platformRepository) Update(ctx context.Context, platform *model.Platform) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		UPDATE platforms 
		SET name = $1, slug = $2, type = $3, thumbnail_url = $4, url_pattern = $5, is_active = $6, is_premium = $7, config = $8
		WHERE id = $9
	`

	args := []interface{}{
		platform.Name, platform.Slug, platform.Type, platform.ThumbnailURL, platform.URLPattern,
		platform.IsActive, platform.IsPremium, platform.Config, platform.ID,
	}
	_, err := r.db.Exec(subCtx, query, args...)
	return err
}

func (r *platformRepository) UpdateThumbnail(ctx context.Context, thumbnail string, platformID uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		UPDATE platforms 
		SET thumbnail_url = $1
		WHERE id = $2
	`

	args := []interface{}{
		thumbnail, platformID,
	}
	_, err := r.db.Exec(subCtx, query, args...)
	return err
}

func (r *platformRepository) Delete(ctx context.Context, id uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `DELETE FROM platforms WHERE id = $1`
	args := []interface{}{
		id,
	}
	cmdTag, err := r.db.Exec(subCtx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete platform: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("platform with id %s not found", id)
	}
	return nil
}

func (r *platformRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `DELETE FROM platforms WHERE id = ANY($1)`
	args := []interface{}{
		ids,
	}

	cmdTag, err := r.db.Exec(subCtx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete platforms: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no platforms found with ids %v", ids)
	}
	return nil
}
