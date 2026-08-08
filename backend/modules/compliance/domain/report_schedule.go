package domain

import (
	"time"

	"github.com/google/uuid"
)

const DefaultWindowDays = 30

type ReportSchedule struct {
	ID                uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID          uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index"`
	UserID            uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	FrameworkKey      string    `gorm:"column:framework_key;size:100;not null"`
	ScheduleString    string    `gorm:"column:schedule_string;size:250;not null"` // 5-field cron
	WindowDays        int       `gorm:"column:window_days;not null;default:30"`
	To                string    `gorm:"column:to_addresses;size:2000"` // comma-separated
	Cc                string    `gorm:"column:cc_addresses;size:2000"` // comma-separated
	LastExecutionDate time.Time `gorm:"column:last_execution_date;not null"`
	NextExecutionDate time.Time `gorm:"column:next_execution_date;not null;index"`
}

func (ReportSchedule) TableName() string { return "compliance_report_schedule" }
