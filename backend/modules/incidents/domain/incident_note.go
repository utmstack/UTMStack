package domain

import (
	"time"

	"github.com/google/uuid"
)

type IncidentNote struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"                                         json:"id"`
	TenantID     uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index:idx_incident_note_tenant_incident,priority:1" json:"-"`
	IncidentID   uuid.UUID `gorm:"column:incident_id;type:uuid;not null;index:idx_incident_note_tenant_incident,priority:2" json:"incidentId"`
	NoteText     string    `gorm:"column:note_text;size:1000;not null"                                                    json:"noteText"`
	NoteSendDate time.Time `gorm:"column:note_send_date;not null;default:now()"                                           json:"noteSendDate"`
	NoteSendBy   *string   `gorm:"column:note_send_by;size:255"                                                           json:"noteSendBy,omitempty"`
}

func (IncidentNote) TableName() string { return "incident_note" }
