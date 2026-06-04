package domain

import "time"

type UtmIncidentNote struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"                json:"id"`
	IncidentID   int64     `gorm:"column:incident_id;not null"             json:"incidentId"`
	NoteText     string    `gorm:"column:note_text;size:1000;not null"     json:"noteText"`
	NoteSendDate time.Time `gorm:"column:note_send_date;not null;default:now()" json:"noteSendDate"`
	NoteSendBy   *string   `gorm:"column:note_send_by;size:255"            json:"noteSendBy,omitempty"`
}

func (UtmIncidentNote) TableName() string { return "utm_incident_note" }
