package domain

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"column:user_id;type:uuid;not null;index;uniqueIndex:ux_api_key_owner_name,priority:1"`
	TenantID    uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null;index"`
	Name        string     `gorm:"size:255;not null;uniqueIndex:ux_api_key_owner_name,priority:2"`
	KeyHash     string     `gorm:"column:key_hash;size:64;not null;uniqueIndex"`
	AllowedIP   string     `gorm:"column:allowed_ip;type:text"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	GeneratedAt time.Time  `gorm:"column:generated_at;not null"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
}

func (APIKey) TableName() string { return "api_keys" }
