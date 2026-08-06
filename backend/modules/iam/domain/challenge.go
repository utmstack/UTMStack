package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChallengePurpose string

const (
	ChallengeActivation    ChallengePurpose = "activation"
	ChallengePasswordReset ChallengePurpose = "password_reset"
	ChallengeTfaEnrollment ChallengePurpose = "tfa_enrollment"
	ChallengeTfaLogin      ChallengePurpose = "tfa_login"
)

type UserChallenge struct {
	UserID     uuid.UUID        `gorm:"column:user_id;type:uuid;primaryKey"`
	Purpose    ChallengePurpose `gorm:"column:purpose;primaryKey;size:32"`
	TenantID   uuid.UUID        `gorm:"column:tenant_id;type:uuid;not null;index"`
	FactorID   *uuid.UUID       `gorm:"column:factor_id;type:uuid;index"`
	Secret     string           `gorm:"column:secret;type:text"`
	LastCode   string           `gorm:"column:last_code;size:16"`
	ExpiresAt  time.Time        `gorm:"column:expires_at"`
	CooldownAt time.Time        `gorm:"column:cooldown_at"`
}

func (UserChallenge) TableName() string { return "user_challenge" }
