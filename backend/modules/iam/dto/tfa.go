package dto

import "time"

type TfaInitRequest struct {
	Method string `json:"method" binding:"required,oneof=EMAIL TOTP"`
}

type TfaInitResponse struct {
	Method     string    `json:"method"`
	QRDataURL  string    `json:"qr_data_url,omitempty"`
	OtpAuthURL string    `json:"otp_auth_url,omitempty"`
	EmailSent  bool      `json:"email_sent,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type TfaVerifyRequest struct {
	Method string `json:"method" binding:"required,oneof=EMAIL TOTP"`
	Code   string `json:"code" binding:"required,len=6"`
}

type TfaCompleteRequest struct {
	Method string `json:"method" binding:"required,oneof=EMAIL TOTP"`
}

type TfaRefreshRequest struct {
	Method string `json:"method" binding:"required,oneof=EMAIL TOTP"`
}

type TfaRefreshResponse struct {
	EmailSent     bool      `json:"email_sent"`
	ExpiresAt     time.Time `json:"expires_at"`
	CooldownUntil time.Time `json:"cooldown_until,omitempty"`
}

type TfaVerifyCodeRequest struct {
	PreAuthToken string `json:"pre_auth_token" binding:"required"`
	Code         string `json:"code" binding:"required,len=6"`
}

type TfaEnrollmentRequest struct {
	Stage  string `json:"stage" binding:"required,oneof=INIT VERIFY COMPLETE"`
	Method string `json:"method" binding:"required,oneof=EMAIL TOTP"`
	Code   string `json:"code,omitempty"`
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
