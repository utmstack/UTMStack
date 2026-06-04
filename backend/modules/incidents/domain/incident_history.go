package domain

import "time"

type UtmIncidentHistory struct {
	ID                int64     `gorm:"primaryKey;autoIncrement"                        json:"id"`
	IncidentID        int64     `gorm:"column:incident_id;not null"                     json:"incidentId"`
	Action            string    `gorm:"column:action;size:255"                          json:"action"`
	ActionType        string    `gorm:"column:action_type;size:100;not null"            json:"actionType"`
	ActionDetail      *string   `gorm:"column:action_detail;type:text"                  json:"actionDetail,omitempty"`
	ActionCreatedDate time.Time `gorm:"column:action_created_date;not null;default:now()" json:"actionCreatedDate"`
	ActionCreatedBy   *string   `gorm:"column:action_created_by;size:255"               json:"actionCreatedBy,omitempty"`
}

func (UtmIncidentHistory) TableName() string { return "utm_incident_history" }
