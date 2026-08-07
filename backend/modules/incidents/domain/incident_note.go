package domain

import (
	"time"

	"github.com/google/uuid"
)

type UtmIncidentNote struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"                json:"id"`
	TenantID     uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index" json:"-"`
	IncidentID   int64     `gorm:"column:incident_id;not null"             json:"incidentId"`
	NoteText     string    `gorm:"column:note_text;size:1000;not null"     json:"noteText"`
	NoteSendDate time.Time `gorm:"column:note_send_date;not null;default:now()" json:"noteSendDate"`
	NoteSendBy   *string   `gorm:"column:note_send_by;size:255"            json:"noteSendBy,omitempty"`
}

func (UtmIncidentNote) TableName() string { return "incident_note" }
