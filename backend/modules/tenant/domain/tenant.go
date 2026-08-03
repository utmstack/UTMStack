package domain

import "time"

type TenantStatus string

const (
	StatusActive     TenantStatus = "ACTIVE"
	StatusSuspended  TenantStatus = "SUSPENDED"
	StatusTerminated TenantStatus = "TERMINATED"
)

type Tenant struct {
	ID        string       `gorm:"primaryKey;size:36" json:"id"`
	Name      string       `gorm:"size:255;not null" json:"name"`
	Domain    string       `gorm:"size:253;not null;uniqueIndex" json:"domain"`
	Status    TenantStatus `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	CreatedAt time.Time    `json:"createdAt"`
}

func (Tenant) TableName() string { return "tenant" }
