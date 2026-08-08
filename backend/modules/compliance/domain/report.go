package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ComplianceStatus string

const (
	StatusCompliant    ComplianceStatus = "COMPLIANT"     // evaluated and passing
	StatusNonCompliant ComplianceStatus = "NON_COMPLIANT" // evaluated and failing
	StatusAtRisk       ComplianceStatus = "AT_RISK"       // passing analysis but weak coverage (or vice-versa)
	StatusNotCovered   ComplianceStatus = "NOT_COVERED"   // nothing is watching this: no checks and no covering rules
	StatusNotEvaluated ComplianceStatus = "NOT_EVALUATED" // checks exist but the tenant receives no such data
	StatusPending      ComplianceStatus = "PENDING"       // check declared but not yet written
	StatusOutOfScope   ComplianceStatus = "OUT_OF_SCOPE"  // governance control — not provable from logs
)

type CheckOutcome string

const (
	CheckPassed        CheckOutcome = "PASSED"
	CheckFailed        CheckOutcome = "FAILED"
	CheckNotApplicable CheckOutcome = "NOT_APPLICABLE" // the tenant receives no data of this type
	CheckError         CheckOutcome = "ERROR"          // the check could not run, which is not the same as failing it
)

type Report struct {
	ID              uuid.UUID      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID      `gorm:"column:tenant_id;type:uuid;not null;uniqueIndex:uniq_compliance_report_tenant_fw,priority:1"`
	FrameworkKey    string         `gorm:"column:framework_key;size:100;not null;uniqueIndex:uniq_compliance_report_tenant_fw,priority:2"`
	FrameworkName   string         `gorm:"column:framework_name;size:250;not null"`
	FrameworkSource string         `gorm:"column:framework_source;size:250"`
	GeneratedAt     time.Time      `gorm:"column:generated_at;not null"`
	WindowFrom      time.Time      `gorm:"column:window_from;not null"`
	WindowTo        time.Time      `gorm:"column:window_to;not null"`
	Score           int            `gorm:"column:score;not null"`
	Version         int            `gorm:"column:version;not null;default:0"`
	Body            datatypes.JSON `gorm:"column:body;type:jsonb;not null"`
}

func (Report) TableName() string { return "compliance_report" }
