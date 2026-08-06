package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	TenantID  uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index"`
	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
	CreatedAt time.Time `gorm:"not null"`
	IP        string    `gorm:"size:45"`
	UserAgent string    `gorm:"size:255"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
