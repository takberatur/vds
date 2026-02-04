package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
)

type ApplicationRepository interface {
	BaseRepository
	Create(ctx context.Context, app *model.Application) error
	FindByPackageName(ctx context.Context, packageName string) (*model.Application, error)
	FindByAPIKey(ctx context.Context, apiKey string) (*model.Application, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Application, error)
	Update(ctx context.Context, app *model.Application) error
	Delete(ctx context.Context, id uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	FindAll(ctx context.Context, params model.QueryParamsRequest) ([]model.Application, model.Pagination, error)
	GetAll(ctx context.Context) ([]*model.Application, error)
}

type applicationRepository struct {
	*baseRepository
}

func NewApplicationRepository(db *pgxpool.Pool) ApplicationRepository {
	return &applicationRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}

func (r *applicationRepository) Create(ctx context.Context, app *model.Application) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		INSERT INTO applications (name, package_name, api_key, secret_key, version, platform, enable_monetization, enable_admob, enable_unity_ad, enable_start_app, enable_in_app_purchase, admob_ad_unit_id, unity_ad_unit_id, start_app_ad_unit_id,
		admob_banner_ad_unit_id, admob_interstitial_ad_unit_id, admob_native_ad_unit_id, admob_rewarded_ad_unit_id,
		unity_banner_ad_unit_id, unity_interstitial_ad_unit_id, unity_native_ad_unit_id, unity_rewarded_ad_unit_id, 
		one_signal_id,
		is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRow(subCtx, query,
		app.Name,
		app.PackageName,
		app.APIKey,
		app.SecretKey,
		app.Version,
		app.Platform,
		app.EnableMonetization,
		app.EnableAdmob,
		app.EnableUnityAd,
		app.EnableStartApp,
		app.EnableInAppPurchase,
		app.AdmobAdUnitID,
		app.UnityAdUnitID,
		app.StartAppAdUnitID,
		app.AdmobBannerAdUnitID,
		app.AdmobInterstitialAdUnitID,
		app.AdmobNativeAdUnitID,
		app.AdmobRewardedAdUnitID,
		app.UnityBannerAdUnitID,
		app.UnityInterstitialAdUnitID,
		app.UnityNativeAdUnitID,
		app.UnityRewardedAdUnitID,
		app.OneSignalID,
		true, // is_active default
		now,
		now,
	).Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)

	return err
}

func (r *applicationRepository) FindByPackageName(ctx context.Context, packageName string) (*model.Application, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `SELECT id, name, package_name, api_key, secret_key, version, platform, enable_monetization, enable_admob, enable_unity_ad, enable_start_app, enable_in_app_purchase, admob_ad_unit_id, unity_ad_unit_id, start_app_ad_unit_id,
		admob_banner_ad_unit_id, admob_interstitial_ad_unit_id, admob_native_ad_unit_id, admob_rewarded_ad_unit_id,
		unity_banner_ad_unit_id, unity_interstitial_ad_unit_id, unity_native_ad_unit_id, unity_rewarded_ad_unit_id,
		one_signal_id,
		is_active, created_at, updated_at FROM applications WHERE package_name = $1`

	var app model.Application
	err := r.db.QueryRow(subCtx, query, packageName).Scan(
		&app.ID, &app.Name, &app.PackageName, &app.APIKey, &app.SecretKey,
		&app.Version, &app.Platform, &app.EnableMonetization, &app.EnableAdmob, &app.EnableUnityAd, &app.EnableStartApp, &app.EnableInAppPurchase,
		&app.AdmobAdUnitID, &app.UnityAdUnitID, &app.StartAppAdUnitID,
		&app.AdmobBannerAdUnitID, &app.AdmobInterstitialAdUnitID, &app.AdmobNativeAdUnitID, &app.AdmobRewardedAdUnitID,
		&app.UnityBannerAdUnitID, &app.UnityInterstitialAdUnitID, &app.UnityNativeAdUnitID, &app.UnityRewardedAdUnitID,
		&app.OneSignalID,
		&app.IsActive, &app.CreatedAt, &app.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &app, nil
}

func (r *applicationRepository) FindByAPIKey(ctx context.Context, apiKey string) (*model.Application, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `SELECT id, name, package_name, api_key, secret_key, version, platform, enable_monetization, enable_admob, enable_unity_ad, enable_start_app, enable_in_app_purchase, admob_ad_unit_id, unity_ad_unit_id, start_app_ad_unit_id,
		admob_banner_ad_unit_id, admob_interstitial_ad_unit_id, admob_native_ad_unit_id, admob_rewarded_ad_unit_id,
		unity_banner_ad_unit_id, unity_interstitial_ad_unit_id, unity_native_ad_unit_id, unity_rewarded_ad_unit_id,
		one_signal_id,
		is_active, created_at, updated_at FROM applications WHERE api_key = $1`

	var app model.Application
	err := r.db.QueryRow(subCtx, query, apiKey).Scan(
		&app.ID, &app.Name, &app.PackageName, &app.APIKey, &app.SecretKey,
		&app.Version, &app.Platform, &app.EnableMonetization, &app.EnableAdmob, &app.EnableUnityAd, &app.EnableStartApp, &app.EnableInAppPurchase,
		&app.AdmobAdUnitID, &app.UnityAdUnitID, &app.StartAppAdUnitID,
		&app.AdmobBannerAdUnitID, &app.AdmobInterstitialAdUnitID, &app.AdmobNativeAdUnitID, &app.AdmobRewardedAdUnitID,
		&app.UnityBannerAdUnitID, &app.UnityInterstitialAdUnitID, &app.UnityNativeAdUnitID, &app.UnityRewardedAdUnitID,
		&app.OneSignalID,
		&app.IsActive, &app.CreatedAt, &app.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &app, nil
}

func (r *applicationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Application, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `SELECT id, name, package_name, api_key, secret_key, version, platform, enable_monetization, enable_admob, enable_unity_ad, enable_start_app, enable_in_app_purchase, admob_ad_unit_id, unity_ad_unit_id, start_app_ad_unit_id,
		admob_banner_ad_unit_id, admob_interstitial_ad_unit_id, admob_native_ad_unit_id, admob_rewarded_ad_unit_id,
		unity_banner_ad_unit_id, unity_interstitial_ad_unit_id, unity_native_ad_unit_id, unity_rewarded_ad_unit_id,
		one_signal_id,
		is_active, created_at, updated_at FROM applications WHERE id = $1`

	var app model.Application
	err := r.db.QueryRow(subCtx, query, id).Scan(
		&app.ID, &app.Name, &app.PackageName, &app.APIKey, &app.SecretKey,
		&app.Version, &app.Platform, &app.EnableMonetization, &app.EnableAdmob, &app.EnableUnityAd, &app.EnableStartApp, &app.EnableInAppPurchase,
		&app.AdmobAdUnitID, &app.UnityAdUnitID, &app.StartAppAdUnitID,
		&app.AdmobBannerAdUnitID, &app.AdmobInterstitialAdUnitID, &app.AdmobNativeAdUnitID, &app.AdmobRewardedAdUnitID,
		&app.UnityBannerAdUnitID, &app.UnityInterstitialAdUnitID, &app.UnityNativeAdUnitID, &app.UnityRewardedAdUnitID,
		&app.OneSignalID,
		&app.IsActive, &app.CreatedAt, &app.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &app, nil
}

func (r *applicationRepository) Update(ctx context.Context, app *model.Application) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		UPDATE applications SET
			name = $1, package_name = $2, version = $3, platform = $4,
			enable_monetization = $5, enable_admob = $6, enable_unity_ad = $7, enable_start_app = $8, enable_in_app_purchase = $9,
			admob_ad_unit_id = $10, unity_ad_unit_id = $11, start_app_ad_unit_id = $12,
			admob_banner_ad_unit_id = $13, admob_interstitial_ad_unit_id = $14, admob_native_ad_unit_id = $15, admob_rewarded_ad_unit_id = $16,
			unity_banner_ad_unit_id = $17, unity_interstitial_ad_unit_id = $18, unity_native_ad_unit_id = $19, unity_rewarded_ad_unit_id = $20,
			one_signal_id = $21,
			is_active = $22, updated_at = $23
		WHERE id = $24
	`

	_, err := r.db.Exec(subCtx, query,
		app.Name, app.PackageName, app.Version, app.Platform,
		app.EnableMonetization, app.EnableAdmob, app.EnableUnityAd, app.EnableStartApp, app.EnableInAppPurchase,
		app.AdmobAdUnitID, app.UnityAdUnitID, app.StartAppAdUnitID,
		app.AdmobBannerAdUnitID, app.AdmobInterstitialAdUnitID, app.AdmobNativeAdUnitID, app.AdmobRewardedAdUnitID,
		app.UnityBannerAdUnitID, app.UnityInterstitialAdUnitID, app.UnityNativeAdUnitID, app.UnityRewardedAdUnitID,
		app.OneSignalID,
		app.IsActive, time.Now(), app.ID,
	)

	return err
}

func (r *applicationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `DELETE FROM applications WHERE id = $1`
	_, err := r.db.Exec(subCtx, query, id)
	return err
}

func (r *applicationRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `DELETE FROM applications WHERE id = ANY($1)`
	_, err := r.db.Exec(subCtx, query, ids)
	return err
}

func (r *applicationRepository) FindAll(ctx context.Context, params model.QueryParamsRequest) ([]model.Application, model.Pagination, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	qb := NewQueryBuilder(`SELECT id, name, package_name, api_key, secret_key, version, platform, enable_monetization, enable_admob, enable_unity_ad, enable_start_app, enable_in_app_purchase, admob_ad_unit_id, unity_ad_unit_id, start_app_ad_unit_id,
		admob_banner_ad_unit_id, admob_interstitial_ad_unit_id, admob_native_ad_unit_id, admob_rewarded_ad_unit_id,
		unity_banner_ad_unit_id, unity_interstitial_ad_unit_id, unity_native_ad_unit_id, unity_rewarded_ad_unit_id,
		one_signal_id,
		is_active, created_at, updated_at FROM applications`)

	if params.Search != "" {
		qb.Where("(name ILIKE $? OR package_name ILIKE $?)", "%"+params.Search+"%", "%"+params.Search+"%")
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

	// Sorting
	if params.SortBy != "" {
		qb.OrderByField(params.SortBy, params.OrderBy)
	} else {
		qb.OrderByField("created_at", "DESC")
	}

	// Count Total (before limit/offset)
	countQuery, countArgs := qb.Clone().ChangeBase("SELECT COUNT(*) FROM applications").WithoutPagination().Build()

	var totalItems int64
	err := r.db.QueryRow(subCtx, countQuery, countArgs...).Scan(&totalItems)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("failed to count applications: %w", err)
	}

	// Pagination
	offset := (params.Page - 1) * params.Limit
	qb.WithLimit(params.Limit).WithOffset(offset)

	query, args := qb.Build()
	rows, err := r.db.Query(subCtx, query, args...)
	if err != nil {
		return nil, model.Pagination{}, err
	}
	defer rows.Close()

	var apps []model.Application
	for rows.Next() {
		var app model.Application
		if err := rows.Scan(
			&app.ID, &app.Name, &app.PackageName, &app.APIKey, &app.SecretKey,
			&app.Version, &app.Platform, &app.EnableMonetization, &app.EnableAdmob, &app.EnableUnityAd, &app.EnableStartApp, &app.EnableInAppPurchase,
			&app.AdmobAdUnitID, &app.UnityAdUnitID, &app.StartAppAdUnitID,
			&app.AdmobBannerAdUnitID, &app.AdmobInterstitialAdUnitID, &app.AdmobNativeAdUnitID, &app.AdmobRewardedAdUnitID,
			&app.UnityBannerAdUnitID, &app.UnityInterstitialAdUnitID, &app.UnityNativeAdUnitID, &app.UnityRewardedAdUnitID,
			&app.OneSignalID,
			&app.IsActive, &app.CreatedAt, &app.UpdatedAt,
		); err != nil {
			return nil, model.Pagination{}, err
		}
		apps = append(apps, app)
	}

	pagination := model.Pagination{
		CurrentPage: params.Page,
		Limit:       params.Limit,
		TotalItems:  totalItems,
		TotalPages:  int((totalItems + int64(params.Limit) - 1) / int64(params.Limit)),
		HasNext:     int64(params.Page*params.Limit) < totalItems,
		HasPrev:     params.Page > 1,
	}

	return apps, pagination, nil
}

func (r *applicationRepository) GetAll(ctx context.Context) ([]*model.Application, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `SELECT * FROM applications`

	var apps []*model.Application
	err := pgxscan.Select(subCtx, r.db, &apps, query)
	if err != nil {
		return nil, err
	}

	return apps, nil
}
