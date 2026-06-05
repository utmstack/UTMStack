package domain

type UtmAssetTypes struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement;column:id"`
	TypeName string `gorm:"column:type_name;size:100;uniqueIndex"`
}

func (UtmAssetTypes) TableName() string { return "utm_asset_types" }
