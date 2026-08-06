package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
)

type TfaInitResponse struct {
	Type       domain.TfaFactorType `json:"type"`
	FactorID   uuid.UUID            `json:"factor_id"`
	QRDataURL  string               `json:"qr_data_url,omitempty"`
	OtpAuthURL string               `json:"otp_auth_url,omitempty"`
	EmailSent  bool                 `json:"email_sent,omitempty"`
	ExpiresAt  time.Time            `json:"expires_at"`
}

type TfaVerifyCodeRequest struct {
	PreAuthToken string `json:"pre_auth_token" binding:"required"`
	Code         string `json:"code" binding:"required,len=6"`
}

type TfaEnrollmentRequest struct {
	Stage string               `json:"stage" binding:"required,oneof=INIT VERIFY COMPLETE"`
	Type  domain.TfaFactorType `json:"type" binding:"required,oneof=email totp"`
	Code  string               `json:"code,omitempty"`
}

type TfaEnrollmentResponse struct {
	Stage    string           `json:"stage"`
	Init     *TfaInitResponse `json:"init,omitempty"`
	Verified *bool            `json:"verified,omitempty"`
	Enabled  *bool            `json:"enabled,omitempty"`
}

type TfaDisableRequest struct {
	Password string `json:"password" binding:"required"`
}

type TfaFactorResponse struct {
	ID          uuid.UUID            `json:"id"`
	Type        domain.TfaFactorType `json:"type"`
	ConfirmedAt *time.Time           `json:"confirmed_at,omitempty"`
	LastUsedAt  *time.Time           `json:"last_used_at,omitempty"`
}
