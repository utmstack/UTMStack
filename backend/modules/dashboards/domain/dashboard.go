package domain

import "time"

type Dashboard struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"column:name;size:100;not null;uniqueIndex" json:"name"`
	Description  string    `gorm:"column:description;size:255" json:"description"`
	Config       string    `gorm:"column:config" json:"config"`
	// RefreshTime is the auto-refresh interval in seconds. 0 disables auto-refresh.
	RefreshTime  int64     `gorm:"column:refresh_time;not null;default:0" json:"refreshTime"`
	SystemOwner  bool      `gorm:"column:system_owner" json:"systemOwner"`
	CreatedDate  time.Time `gorm:"column:created_date" json:"createdDate"`
	ModifiedDate time.Time `gorm:"column:modified_date" json:"modifiedDate"`
}

func (Dashboard) TableName() string { return "dashboard" }
