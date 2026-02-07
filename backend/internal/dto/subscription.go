package dto

type MobileSubscriptionUpsertRequest struct {
	OriginalTransactionID string `json:"original_transaction_id" validate:"required"`
	ProductID             string `json:"product_id" validate:"required"`
	PurchaseToken         string `json:"purchase_token" validate:"required"`
	Platform              string `json:"platform" validate:"omitempty"`
	StartTimeMs           int64  `json:"start_time_ms" validate:"required"`
	EndTimeMs             int64  `json:"end_time_ms" validate:"required"`
	Status                string `json:"status" validate:"omitempty,oneof=active expired canceled"`
	AutoRenew             bool   `json:"auto_renew" validate:"omitempty"`
}

