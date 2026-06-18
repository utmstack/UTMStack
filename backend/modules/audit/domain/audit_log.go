package domain

import (
	"time"

	"gorm.io/datatypes"
)

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
)

type AuditLog struct {
	ID           uint64               `gorm:"primaryKey;autoIncrement"`
	Timestamp    time.Time            `gorm:"not null;index;default:CURRENT_TIMESTAMP"`
	UserLogin    string               `gorm:"size:50;index"`
	ErrorMessage string               `gorm:"type:text"`
	UserID       *uint64              `gorm:"index"`
	IP           string               `gorm:"size:45"`
	UserAgent    string               `gorm:"size:255"`
	SessionID    *uint64              `gorm:"index"`
	Action       string               `gorm:"size:64;index"`
	Status       string               `gorm:"size:16;index"` // success | failure
	EventType    ApplicationEventType `gorm:"size:64;not null;index"`
	ResourceType string               `gorm:"size:32;index"`
	ResourceID   string               `gorm:"size:128;index"`
	Metadata     datatypes.JSON       `gorm:"type:jsonb"`
}

func (AuditLog) TableName() string { return "audit_logs" }
