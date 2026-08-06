package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	TfaStageInit     = "INIT"
	TfaStageVerify   = "VERIFY"
	TfaStageComplete = "COMPLETE"
)

type TfaFactorType string

const (
	TfaFactorEmail    TfaFactorType = "email"
	TfaFactorTotp     TfaFactorType = "totp"
	TfaFactorRecovery TfaFactorType = "recovery"
)

type TfaFactor struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID     `gorm:"column:tenant_id;type:uuid;not null;index"`
	UserID      uuid.UUID     `gorm:"column:user_id;type:uuid;not null;index:ix_tfa_factor_user,priority:1"`
	Type        TfaFactorType `gorm:"size:16;not null;index:ix_tfa_factor_user,priority:2"`
	Secret      string        `gorm:"type:text"`
	ConfirmedAt *time.Time
	LastUsedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (TfaFactor) TableName() string { return "user_tfa_factor" }
