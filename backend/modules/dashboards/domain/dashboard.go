package domain

import (
	"time"

	"github.com/google/uuid"
)

type Dashboard struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID     uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index;uniqueIndex:idx_dashboard_tenant_name,priority:1" json:"-"`
	Name         string    `gorm:"column:name;size:100;not null;uniqueIndex:idx_dashboard_tenant_name,priority:2" json:"name"`
	Description  string    `gorm:"column:description;size:255" json:"description"`
	Config       string    `gorm:"column:config" json:"config"`
	SystemOwner  bool      `gorm:"column:system_owner;not null;default:false" json:"systemOwner"`
	CreatedDate  time.Time `gorm:"column:created_date" json:"createdDate"`
	ModifiedDate time.Time `gorm:"column:modified_date" json:"modifiedDate"`
}

func (Dashboard) TableName() string { return "dashboard" }

func (Dashboard) SystemFlagColumn() string { return "system_owner" }
