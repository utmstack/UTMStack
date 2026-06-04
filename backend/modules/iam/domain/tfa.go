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

type TfaSetupState struct {
	UserID     uint64
	Purpose    string
	Method     string
	Secret     string
	ExpiresAt  time.Time
	CooldownAt time.Time
	LastCode   string
	Verified   bool
}
