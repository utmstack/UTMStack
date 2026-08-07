package domain

import (
	"time"

	"gorm.io/datatypes"
)

type UtmComplianceReportSnapshot struct {
	ID            string         `gorm:"column:id;primaryKey;size:36"`
	TenantID      string         `gorm:"column:tenant_id;size:36;index;not null"`
	FrameworkKey  string         `gorm:"column:framework_key;size:100;not null;index"`
	FrameworkName string         `gorm:"column:framework_name;size:250;not null"`
	Timestamp     time.Time      `gorm:"column:generated_at;not null;index:idx_report_snapshot_time,priority:2"`
	Score         int            `gorm:"column:score;not null"`
	Report        datatypes.JSON `gorm:"column:report;type:jsonb;not null"`
}

func (UtmComplianceReportSnapshot) TableName() string { return "compliance_report_snapshot" }
