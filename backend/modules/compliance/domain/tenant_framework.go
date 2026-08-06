package domain

import "time"

type TenantFramework struct {
	TenantID     string    `gorm:"column:tenant_id;size:36;primaryKey"`
	FrameworkKey string    `gorm:"column:framework_key;size:100;primaryKey"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (TenantFramework) TableName() string { return "compliance_tenant_framework" }
