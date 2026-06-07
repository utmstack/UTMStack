package domain

import "time"

type UtmAssetGroup struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	GroupName        string    `gorm:"column:group_name;size:100;not null;uniqueIndex"`
	GroupDescription string    `gorm:"column:group_description;size:255"`
	CreatedDate      time.Time `gorm:"column:created_date"`
}

func (UtmAssetGroup) TableName() string { return "utm_asset_group" }
