package domain

import "time"

const (
	TfaMethodEmail = "EMAIL"
	TfaMethodTotp  = "TOTP"

	TfaPurposeEnrollment = "ENROLLMENT"
	TfaPurposeLogin      = "LOGIN"

	TfaStageInit     = "INIT"
	TfaStageVerify   = "VERIFY"
	TfaStageComplete = "COMPLETE"
)

// TfaSetupState is a short-lived 2FA enrollment/login challenge. Persisted in
// Postgres (not just in-memory) so a horizontally-scaled backend's replicas
// share the same state — otherwise a setup started against one replica and
// completed against another (different LB pick) would never find it.
type TfaSetupState struct {
	UserID     uint64    `gorm:"column:user_id;primaryKey"`
	Purpose    string    `gorm:"column:purpose;primaryKey;size:32"`
	Method     string    `gorm:"column:method;primaryKey;size:16"`
	Secret     string    `gorm:"column:secret;type:text"`
	ExpiresAt  time.Time `gorm:"column:expires_at"`
	CooldownAt time.Time `gorm:"column:cooldown_at"`
	LastCode   string    `gorm:"column:last_code;size:16"`
	Verified   bool      `gorm:"column:verified;not null;default:false"`
}

func (TfaSetupState) TableName() string { return "utm_tfa_setup_state" }
