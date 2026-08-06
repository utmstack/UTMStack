package domain

import "time"

type UtmAssetGroup struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID         string    `gorm:"column:tenant_id;size:36;index;uniqueIndex:idx_asset_group_tenant_name" json:"-"`
	GroupName        string    `gorm:"column:group_name;size:100;not null;uniqueIndex:idx_asset_group_tenant_name"`
	GroupDescription string    `gorm:"column:group_description;size:255"`
	CreatedDate      time.Time `gorm:"column:created_date"`
}

func (UtmAssetGroup) TableName() string { return "utm_asset_group" }
