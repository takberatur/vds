package dto

type ContactRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type WebErrorReport struct {
	Error       string `json:"error" validate:"omitempty"`
	Message     string `json:"message" validate:"omitempty"`
	PlatformID  string `json:"platform_id" validate:"omitempty"`
	UserID      string `json:"user_id" validate:"omitempty"`
	IPAddress   string `json:"ip_address" validate:"omitempty"`
	UserAgent   string `json:"user_agent" validate:"omitempty"`
	URL         string `json:"url" validate:"omitempty"`
	Method      string `json:"method" validate:"omitempty"`
	Request     string `json:"request" validate:"omitempty"`
	Status      int    `json:"status" validate:"omitempty"`
	Level       string `json:"level" validate:"omitempty"`
	Locale      string `json:"locale" validate:"omitempty"`
	TimestampMs int64  `json:"timestamp_ms" validate:"omitempty"`
}
