package domain

import (
	"time"

	"github.com/google/uuid"
)

type Action string

const (
	ActionCreated            Action = "INCIDENT_CREATED"
	ActionAlertAdd           Action = "INCIDENT_ALERT_ADD"
	ActionAlertStatusChanged Action = "INCIDENT_ALERT_STATUS_CHANGED"
	ActionStatusChange       Action = "INCIDENT_STATUS_CHANGE"
	ActionNoteAdd            Action = "INCIDENT_NOTE_ADD"
	ActionNoteChange         Action = "INCIDENT_NOTE_CHANGE"
	ActionAssigned           Action = "INCIDENT_ASSIGNED"
)

type IncidentHistory struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"                                            json:"id"`
	TenantID          uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index:idx_incident_history_tenant_incident,priority:1" json:"-"`
	IncidentID        uuid.UUID `gorm:"column:incident_id;type:uuid;not null;index:idx_incident_history_tenant_incident,priority:2" json:"incidentId"`
	Action            Action    `gorm:"column:action;type:varchar(64);not null"                                                   json:"action"`
	ActionCreatedDate time.Time `gorm:"column:action_created_date;not null;default:now()"                                         json:"actionCreatedDate"`
	ActionCreatedBy   *string   `gorm:"column:action_created_by;size:255"                                                         json:"actionCreatedBy,omitempty"`
}

func (IncidentHistory) TableName() string { return "incident_history" }
