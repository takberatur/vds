package model

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID                        uuid.UUID `json:"id" db:"id"`
	Name                      string    `json:"name" db:"name"`
	PackageName               string    `json:"package_name" db:"package_name"`
	APIKey                    string    `json:"api_key" db:"api_key"`
	SecretKey                 string    `json:"secret_key" db:"secret_key"`
	Version                   *string   `json:"version" db:"version"`   // Nullable
	Platform                  string    `json:"platform" db:"platform"` // default: android
	EnableMonetization        bool      `json:"enable_monetization" db:"enable_monetization"`
	EnableAdmob               bool      `json:"enable_admob" db:"enable_admob"`
	EnableUnityAd             bool      `json:"enable_unity_ad" db:"enable_unity_ad"`
	EnableStartApp            bool      `json:"enable_start_app" db:"enable_start_app"`
	EnableInAppPurchase       bool      `json:"enable_in_app_purchase" db:"enable_in_app_purchase"`
	AdmobAdUnitID             *string   `json:"admob_ad_unit_id" db:"admob_ad_unit_id"`
	UnityAdUnitID             *string   `json:"unity_ad_unit_id" db:"unity_ad_unit_id"`
	StartAppAdUnitID          *string   `json:"start_app_ad_unit_id" db:"start_app_ad_unit_id"`
	AdmobBannerAdUnitID       *string   `json:"admob_banner_ad_unit_id" db:"admob_banner_ad_unit_id"`
	AdmobInterstitialAdUnitID *string   `json:"admob_interstitial_ad_unit_id" db:"admob_interstitial_ad_unit_id"`
	AdmobNativeAdUnitID       *string   `json:"admob_native_ad_unit_id" db:"admob_native_ad_unit_id"`
	AdmobRewardedAdUnitID     *string   `json:"admob_rewarded_ad_unit_id" db:"admob_rewarded_ad_unit_id"`
	UnityBannerAdUnitID       *string   `json:"unity_banner_ad_unit_id" db:"unity_banner_ad_unit_id"`
	UnityInterstitialAdUnitID *string   `json:"unity_interstitial_ad_unit_id" db:"unity_interstitial_ad_unit_id"`
	UnityNativeAdUnitID       *string   `json:"unity_native_ad_unit_id" db:"unity_native_ad_unit_id"`
	UnityRewardedAdUnitID     *string   `json:"unity_rewarded_ad_unit_id" db:"unity_rewarded_ad_unit_id"`
	OneSignalID               *string   `json:"one_signal_id" db:"one_signal_id"`
	IsActive                  bool      `json:"is_active" db:"is_active"`
	CreatedAt                 time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`

	// Relations
	InAppProducts []InAppProduct `json:"in_app_products,omitempty" db:"-"`
	Subscriptions []Subscription `json:"subscriptions,omitempty" db:"-"`
	Downloads     []DownloadTask `json:"downloads,omitempty" db:"-"`
}

type InAppProduct struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	AppID           uuid.UUID       `json:"app_id" db:"app_id"`
	ProductID       *string         `json:"product_id,omitempty" db:"product_id"`
	ProductType     *string         `json:"product_type,omitempty" db:"product_type"`
	SkuCode         *string         `json:"sku_code,omitempty" db:"sku_code"`
	Title           *string         `json:"title,omitempty" db:"title"`
	Description     *string         `json:"description,omitempty" db:"description"`
	Price           *float64        `json:"price,omitempty" db:"price"`
	Currency        *string         `json:"currency,omitempty" db:"currency"`
	BillingPeriod   *string         `json:"billing_period,omitempty" db:"billing_period"`
	TrialPeriodDays *int32          `json:"trial_period_days,omitempty" db:"trial_period_days"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	IsFeatured      bool            `json:"is_featured" db:"is_featured"`
	SortOrder       int32           `json:"sort_order" db:"sort_order"`
	Features        *map[string]any `json:"features,omitempty" db:"features"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type Subscription struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	UserID                *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	AppID                 *uuid.UUID `json:"app_id,omitempty" db:"app_id"`
	OriginalTransactionID string     `json:"original_transaction_id" db:"original_transaction_id"`
	ProductID             string     `json:"product_id" db:"product_id"`
	PurchaseToken         string     `json:"purchase_token" db:"purchase_token"`
	Platform              string     `json:"platform" db:"platform"`
	StartTime             time.Time  `json:"start_time" db:"start_time"`
	EndTime               time.Time  `json:"end_time" db:"end_time"`
	Status                string     `json:"status" db:"status"`
	AutoRenew             bool       `json:"auto_renew" db:"auto_renew"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`

	// Relations
	User         *User         `json:"user,omitempty" db:"-"`
	Transactions []Transaction `json:"transactions,omitempty" db:"-"`
}

type Transaction struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	UserID           *uuid.UUID     `json:"user_id,omitempty" db:"user_id"`
	AppID            *uuid.UUID     `json:"app_id,omitempty" db:"app_id"`
	SubscriptionID   *uuid.UUID     `json:"subscription_id,omitempty" db:"subscription_id"`
	Amount           float64        `json:"amount" db:"amount"`
	Currency         string         `json:"currency" db:"currency"`
	Provider         string         `json:"provider" db:"provider"`
	Status           string         `json:"status" db:"status"`
	ProviderResponse map[string]any `json:"provider_response" db:"provider_response"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`

	// Relations
	User         *User         `json:"user,omitempty" db:"-"`
	Subscription *Subscription `json:"subscription,omitempty" db:"-"`
}

type RegisterAppResponse struct {
	Name        string `json:"name"`
	PackageName string `json:"package_name"`
	APIKey      string `json:"api_key"`
	SecretKey   string `json:"secret_key"`
}

type RegisterAppRequest struct {
	Name                      string  `json:"name" validate:"required"`
	PackageName               string  `json:"package_name" validate:"required"`
	Version                   string  `json:"version"`
	Platform                  string  `json:"platform"`
	EnableMonetization        bool    `json:"enable_monetization"`
	EnableAdmob               bool    `json:"enable_admob"`
	EnableUnityAd             bool    `json:"enable_unity_ad"`
	EnableStartApp            bool    `json:"enable_start_app"`
	EnableInAppPurchase       bool    `json:"enable_in_app_purchase"`
	AdmobAdUnitID             *string `json:"admob_ad_unit_id"`
	UnityAdUnitID             *string `json:"unity_ad_unit_id"`
	StartAppAdUnitID          *string `json:"start_app_ad_unit_id"`
	AdmobBannerAdUnitID       *string `json:"admob_banner_ad_unit_id"`
	AdmobInterstitialAdUnitID *string `json:"admob_interstitial_ad_unit_id"`
	AdmobNativeAdUnitID       *string `json:"admob_native_ad_unit_id"`
	AdmobRewardedAdUnitID     *string `json:"admob_rewarded_ad_unit_id"`
	UnityBannerAdUnitID       *string `json:"unity_banner_ad_unit_id"`
	UnityInterstitialAdUnitID *string `json:"unity_interstitial_ad_unit_id"`
	UnityNativeAdUnitID       *string `json:"unity_native_ad_unit_id"`
	UnityRewardedAdUnitID     *string `json:"unity_rewarded_ad_unit_id"`
	OneSignalID               *string `json:"one_signal_id"`
	IsActive                  bool    `json:"is_active"`
}

type BulkDeleteAppsRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1"`
}
