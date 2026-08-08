package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ReportScore struct {
	TenantID     uuid.UUID      `gorm:"column:tenant_id;type:uuid;not null;primaryKey"`
	FrameworkKey string         `gorm:"column:framework_key;size:100;not null;primaryKey"`
	Day          time.Time      `gorm:"column:day;type:date;not null;primaryKey"`
	GeneratedAt  time.Time      `gorm:"column:generated_at;not null"` // the run this point came from
	Score        int            `gorm:"column:score;not null"`
	Total        int            `gorm:"column:total;not null"`
	Evaluated    int            `gorm:"column:evaluated;not null"`
	Compliant    int            `gorm:"column:compliant;not null"`
	Body         datatypes.JSON `gorm:"column:body;type:jsonb"`
	HasBody      bool           `gorm:"->;column:has_body;-:migration"`
}

func (ReportScore) TableName() string { return "compliance_report_score" }
