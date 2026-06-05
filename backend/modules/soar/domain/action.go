package domain

import "time"

// UtmIncidentAction is a predefined incident-response action (e.g. SHUTDOWN_SERVER,
// RUN_CMD) that operators can run against an agent during incident handling.
type UtmIncidentAction struct {
	ID                int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ActionCommand     *string    `gorm:"column:action_command"              json:"actionCommand"`
	ActionDescription *string    `gorm:"column:action_description"          json:"actionDescription"`
	ActionParams      *string    `gorm:"column:action_params"               json:"actionParams"`
	ActionType        *int       `gorm:"column:action_type"                 json:"actionType"`
	ActionEditable    bool       `gorm:"column:action_editable;not null"    json:"actionEditable"`
	CreatedDate       time.Time  `gorm:"column:created_date;not null"          json:"createdDate"`
	ModifiedDate      *time.Time `gorm:"column:modified_date"                  json:"modifiedDate"`
	CreatedUser       string     `gorm:"column:created_user;size:50;not null"  json:"createdUser"`
	ModifiedUser      *string    `gorm:"column:modified_user;size:50"          json:"modifiedUser"`
}

func (UtmIncidentAction) TableName() string { return "utm_incident_actions" }
