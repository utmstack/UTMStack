package domain

import (
	"time"

	"github.com/google/uuid"
)

type UtmIncidentHistory struct {
	ID                int64     `gorm:"primaryKey;autoIncrement"                        json:"id"`
	TenantID          uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index"       json:"-"`
	IncidentID        int64     `gorm:"column:incident_id;not null"                     json:"incidentId"`
	Action            string    `gorm:"column:action;size:255"                          json:"action"`
	ActionType        string    `gorm:"column:action_type;size:255;not null"            json:"actionType"`
	ActionDetail      *string   `gorm:"column:action_detail;type:text"                  json:"actionDetail,omitempty"`
	ActionCreatedDate time.Time `gorm:"column:action_created_date;not null;default:now()" json:"actionCreatedDate"`
	ActionCreatedBy   *string   `gorm:"column:action_created_by;size:255"               json:"actionCreatedBy,omitempty"`
}

func (UtmIncidentHistory) TableName() string { return "incident_history" }
