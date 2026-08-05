package domain

import "time"

type AIUsage struct {
	TenantID string    `gorm:"column:tenant_id;size:36;not null;primaryKey"`
	Day      time.Time `gorm:"column:day;type:date;not null;primaryKey"`
	Count    int64     `gorm:"column:count;not null;default:0"`
}

func (AIUsage) TableName() string { return "socai_ai_usage" }
