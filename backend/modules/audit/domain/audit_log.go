package domain

import (
	"github.com/google/uuid"
	"time"

	"gorm.io/datatypes"
)

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
)

type AuditLog struct {
	ID            uint64               `gorm:"primaryKey;autoIncrement"`
	TenantID      string               `gorm:"size:36;not null;default:'';index:idx_audit_tenant_id,priority:1;index:idx_audit_tenant_time,priority:1"`
	Timestamp     time.Time            `gorm:"not null;index;default:CURRENT_TIMESTAMP;index:idx_audit_tenant_time,priority:2"`
	UserEmail     string               `gorm:"size:254;index"`
	ErrorMessage  string               `gorm:"type:text"`
	UserID        *uuid.UUID           `gorm:"type:uuid;index"`
	IP            string               `gorm:"size:45"`
	UserAgent     string               `gorm:"size:255"`
	SessionID     *uuid.UUID           `gorm:"type:uuid;index"`
	Action        string               `gorm:"size:64;index"`
	Status        string               `gorm:"size:16;index"` // success | failure
	EventType     ApplicationEventType `gorm:"size:64;not null;index"`
	ResourceType  string               `gorm:"size:32;index"`
	ResourceID    string               `gorm:"size:128;index"`
	Metadata      datatypes.JSON       `gorm:"type:jsonb"`
	SupportAccess string               `gorm:"size:8;index"`
}

func (AuditLog) TableName() string { return "audit_logs" }
