package domain

import (
	"time"

	"github.com/google/uuid"
)

type Visualization struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID     uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index" json:"-"`
	DashboardID  uuid.UUID `gorm:"column:dashboard_id;type:uuid;not null;index" json:"dashboardId"`
	Spec         string    `gorm:"column:spec;type:jsonb" json:"spec"`
	Config       string    `gorm:"column:config" json:"config"`
	Layout       string    `gorm:"column:layout" json:"layout"`
	SystemOwner  bool      `gorm:"column:system_owner;not null;default:false" json:"systemOwner"`
	CreatedDate  time.Time `gorm:"column:created_date" json:"createdDate"`
	ModifiedDate time.Time `gorm:"column:modified_date" json:"modifiedDate"`
}

func (Visualization) TableName() string { return "visualization" }

func (Visualization) SystemFlagColumn() string { return "system_owner" }
