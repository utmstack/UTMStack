package domain

import "time"

// Policy is a centrally-managed EDR policy document. The Doc is the full
// agent-side policy (allowlist, blocklist, sensors, enabled, ...) that a
// policy_set action merges into the endpoint's local config.
type Policy struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;size:36;index" json:"-"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `json:"description"`
	Version     string    `gorm:"size:100;not null" json:"version"`
	Doc         string    `gorm:"type:text;not null" json:"doc"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Policy) TableName() string { return "edr_policies" }
