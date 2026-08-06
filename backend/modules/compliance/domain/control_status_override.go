package domain

import "time"

// UtmComplianceControlStatusOverride is a manual status assignment for a
// (framework, control) pair. Applied on top of the evaluator's computed status
type UtmComplianceControlStatusOverride struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID     string    `gorm:"column:tenant_id;size:36;index;uniqueIndex:ux_ovr_tenant_fw_ctl,priority:1"`
	FrameworkKey string    `gorm:"column:framework_key;size:100;not null;uniqueIndex:ux_ovr_tenant_fw_ctl,priority:2"`
	ControlID    string    `gorm:"column:control_id;size:100;not null;uniqueIndex:ux_ovr_tenant_fw_ctl,priority:3"`
	Status       string    `gorm:"column:status;size:32;not null"`
	Reason       string    `gorm:"column:reason;size:500"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null"`
}

func (UtmComplianceControlStatusOverride) TableName() string {
	return "utm_compliance_control_status_override"
}

func ValidStatus(s string) bool {
	switch s {
	case StatusCompliant, StatusNonCompliant, StatusAtRisk, StatusNotCovered, StatusOutOfScope, StatusPending:
		return true
	}
	return false
}
