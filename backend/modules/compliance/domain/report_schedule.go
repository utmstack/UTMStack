package domain

import "time"

type UtmComplianceReportSchedule struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID            int64     `gorm:"column:user_id;not null"`
	FrameworkKey      string    `gorm:"column:framework_key;size:100;not null"`
	ScheduleString    string    `gorm:"column:schedule_string;size:250;not null"` // 5-field cron
	Recipients        string    `gorm:"column:recipients"`                        // comma-separated emails
	LastExecutionDate time.Time `gorm:"column:last_execution_date;not null"`
}

func (UtmComplianceReportSchedule) TableName() string { return "utm_compliance_report_schedule" }
