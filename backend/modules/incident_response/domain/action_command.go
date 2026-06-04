package domain

type UtmIncidentActionCommand struct {
	ID         int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ActionID   int64   `gorm:"column:action_id;not null"          json:"actionId"`
	OsPlatform *string `gorm:"column:os_platform"                 json:"osPlatform"`
	Command    *string `gorm:"column:command"                     json:"command"`
}

func (UtmIncidentActionCommand) TableName() string { return "utm_incident_action_command" }
