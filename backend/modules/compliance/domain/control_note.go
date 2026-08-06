package domain

import "time"

// UtmComplianceControlNote is a freeform user note attached to a (framework, control)
// pair. Doesn't affect status; surfaced on the report row for the frontend to display.
type UtmComplianceControlNote struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID     string    `gorm:"column:tenant_id;size:36;index;uniqueIndex:ux_note_tenant_fw_ctl,priority:1"`
	FrameworkKey string    `gorm:"column:framework_key;size:100;not null;uniqueIndex:ux_note_tenant_fw_ctl,priority:2"`
	ControlID    string    `gorm:"column:control_id;size:100;not null;uniqueIndex:ux_note_tenant_fw_ctl,priority:3"`
	Note         string    `gorm:"column:note;type:text;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null"`
}

func (UtmComplianceControlNote) TableName() string {
	return "utm_compliance_control_note"
}
