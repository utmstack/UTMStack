package domain

import "time"

type UtmDashboard struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string    `gorm:"column:name;size:100;not null;uniqueIndex" json:"name"`
	Description   string    `gorm:"column:description;size:255" json:"description"`
	RefreshTime   *int64    `gorm:"column:refresh_time" json:"refreshTime,omitempty"`
	CreatedDate   time.Time `gorm:"column:created_date;not null" json:"createdDate"`
	ModifiedDate  time.Time `gorm:"column:modified_date;not null" json:"modifiedDate"`
	UserCreated   string    `gorm:"column:user_created;size:50;not null" json:"userCreated"`
	UserModified  string    `gorm:"column:user_modified;size:50" json:"userModified"`
	Filters       *string   `gorm:"column:filters" json:"filters,omitempty"`
	DashboardType string    `gorm:"column:dashboard_type;size:50" json:"dashboardType"`
	SystemOwner   bool      `gorm:"column:system_owner" json:"systemOwner"`
}

func (UtmDashboard) TableName() string { return "utm_dashboard" }
