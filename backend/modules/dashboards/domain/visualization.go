package domain

import "time"

type Visualization struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;size:36;index" json:"-"`
	DashboardID  uint64    `gorm:"column:dashboard_id;not null;index" json:"dashboardId"`
	Spec         string    `gorm:"column:spec;type:jsonb" json:"spec"`
	Config       string    `gorm:"column:config" json:"config"`
	Layout       string    `gorm:"column:layout" json:"layout"`
	SystemOwner  bool      `gorm:"column:system_owner" json:"systemOwner"`
	CreatedDate  time.Time `gorm:"column:created_date" json:"createdDate"`
	ModifiedDate time.Time `gorm:"column:modified_date" json:"modifiedDate"`
}

func (Visualization) TableName() string { return "visualization" }

func (Visualization) SystemFlagColumn() string { return "system_owner" }
