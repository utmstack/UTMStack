package domain

import "time"

type TenantStatus string

const (
	StatusActive     TenantStatus = "ACTIVE"
	StatusSuspended  TenantStatus = "SUSPENDED"
	StatusTerminated TenantStatus = "TERMINATED"
)

type SupportAccess string

const (
	SupportNone SupportAccess = "NONE"
	SupportRead SupportAccess = "READ"
	SupportFull SupportAccess = "FULL"
)

func (s SupportAccess) Valid() bool {
	return s == SupportNone || s == SupportRead || s == SupportFull
}

type Tenant struct {
	ID            string        `gorm:"primaryKey;size:36" json:"id"`
	Name          string        `gorm:"size:255;not null" json:"name"`
	Domain        string        `gorm:"size:253;not null;uniqueIndex" json:"domain"`
	Status        TenantStatus  `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	SupportAccess SupportAccess `gorm:"size:8;not null;default:NONE" json:"supportAccess"`
	Limits        Limits        `gorm:"embedded;embeddedPrefix:limit_" json:"limits"`
	CreatedAt     time.Time     `json:"createdAt"`
}

type Limits struct {
	MaxAIRequests int `gorm:"not null;default:0" json:"maxAIRequests"`
}

func (l Limits) Valid() bool {
	return l.MaxAIRequests >= 0
}

func (Tenant) TableName() string { return "tenant" }
