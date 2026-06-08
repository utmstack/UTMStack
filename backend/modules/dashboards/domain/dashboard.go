package domain

import "time"

type Dashboard struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"column:name;size:100;not null;uniqueIndex" json:"name"`
	Description  string    `gorm:"column:description;size:255" json:"description"`
	Config       string    `gorm:"column:config" json:"config"`
	SystemOwner  bool      `gorm:"column:system_owner" json:"systemOwner"`
	CreatedDate  time.Time `gorm:"column:created_date" json:"createdDate"`
	ModifiedDate time.Time `gorm:"column:modified_date" json:"modifiedDate"`
}

func (Dashboard) TableName() string { return "dashboard" }
