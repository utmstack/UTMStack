package domain

import "time"

type RefreshToken struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UserID    uint64    `gorm:"index;not null"`
	TenantID  string    `gorm:"column:tenant_id;size:36;not null;default:'';index"`
	TokenHash string    `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
	CreatedAt time.Time `gorm:"not null"`
	IP        string    `gorm:"size:45"`
	UserAgent string    `gorm:"size:255"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

func (t *RefreshToken) IsValid(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return now.Before(t.ExpiresAt)
}
