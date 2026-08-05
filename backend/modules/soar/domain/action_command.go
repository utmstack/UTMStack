package domain

// UtmIncidentActionCommand is the per-OS command backing an incident action.
type UtmIncidentActionCommand struct {
	TenantID   string  `gorm:"column:tenant_id;size:36;not null;default:'';index" json:"-"`
	ID         int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ActionID   int64   `gorm:"column:action_id;not null"          json:"actionId"`
	OsPlatform *string `gorm:"column:os_platform;size:255"        json:"osPlatform"`
	Command    *string `gorm:"column:command"                     json:"command"`
}

func (UtmIncidentActionCommand) TableName() string { return "utm_incident_action_command" }
