package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
)

type SettingRepository interface {
	BaseRepository
	GetAll(ctx context.Context) ([]model.Setting, error)
	GetByGroup(ctx context.Context, groupName string) ([]model.Setting, error)
	GetByKey(ctx context.Context, key string) (*model.Setting, error)
	UpdateByKey(ctx context.Context, key string, value string) error
	// UpdateBulk updates multiple settings at once (useful for saving settings form)
	UpdateBulk(ctx context.Context, settings []model.Setting) error
}

type settingRepository struct {
	*baseRepository
}

func NewSettingRepository(db *pgxpool.Pool) SettingRepository {
	return &settingRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}

func (r *settingRepository) GetAll(ctx context.Context) ([]model.Setting, error) {

	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `SELECT id, key, value, description, group_name, created_at, updated_at FROM settings ORDER BY group_name, key`
	rows, err := r.db.Query(subCtx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []model.Setting
	for rows.Next() {
		var s model.Setting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.Description, &s.GroupName, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *settingRepository) GetByGroup(ctx context.Context, groupName string) ([]model.Setting, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `SELECT id, key, value, description, group_name, created_at, updated_at FROM settings WHERE group_name = $1 ORDER BY key`
	rows, err := r.db.Query(subCtx, query, groupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []model.Setting
	for rows.Next() {
		var s model.Setting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.Description, &s.GroupName, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *settingRepository) GetByKey(ctx context.Context, key string) (*model.Setting, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `SELECT id, key, value, description, group_name, created_at, updated_at FROM settings WHERE key = $1`
	var s model.Setting
	err := r.db.QueryRow(subCtx, query, key).Scan(&s.ID, &s.Key, &s.Value, &s.Description, &s.GroupName, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *settingRepository) UpdateByKey(ctx context.Context, key string, value string) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `UPDATE settings SET value = $1, updated_at = NOW() WHERE key = $2`
	_, err := r.db.Exec(subCtx, query, value, key)
	return err
}

func (r *settingRepository) UpdateBulk(ctx context.Context, settings []model.Setting) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	tx, err := r.db.Begin(subCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(subCtx)

	query := `UPDATE settings SET value = $1, updated_at = NOW() WHERE key = $2`
	for _, s := range settings {
		if _, err := tx.Exec(subCtx, query, s.Value, s.Key); err != nil {
			return fmt.Errorf("failed to update key %s: %w", s.Key, err)
		}
	}

	return tx.Commit(subCtx)
}
