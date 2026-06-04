package domain

type UtmAssetMetric struct {
	ID        string `gorm:"column:id;primaryKey"                                              json:"id"`
	AssetName string `gorm:"column:asset_name;not null;uniqueIndex:uq_asset_metric"            json:"assetName"`
	Metric    string `gorm:"column:metric;not null;uniqueIndex:uq_asset_metric"                json:"metric"`
	Amount    int64  `gorm:"column:amount;not null"                                            json:"amount"`
}

func (UtmAssetMetric) TableName() string { return "utm_asset_metrics" }
