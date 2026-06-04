package domain

import "time"

type APIKey struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement"`
	UserID      uint64     `gorm:"column:user_id;not null;index"`
	Name        string     `gorm:"size:255;not null"`
	APIKey      string     `gorm:"column:api_key;size:255;not null;uniqueIndex"`
	AllowedIP   string     `gorm:"column:allowed_ip;type:text"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	GeneratedAt time.Time  `gorm:"column:generated_at;not null"`
	ExpiresAt   *time.Time `gorm:"column:expires_at"`
}

func (APIKey) TableName() string { return "api_keys" }

func (k *APIKey) Expired(now time.Time) bool {
	return k.ExpiresAt != nil && !k.ExpiresAt.After(now)
}
