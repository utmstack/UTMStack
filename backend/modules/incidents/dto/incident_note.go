package dto

import "time"

type CreateNoteRequest struct {
	IncidentID int64  `json:"incidentId" binding:"required"`
	NoteText   string `json:"noteText"   binding:"required,max=1000"`
}

type UpdateNoteRequest struct {
	ID       int64  `json:"id"       binding:"required"`
	NoteText string `json:"noteText" binding:"required,max=1000"`
}

type NoteResponse struct {
	ID           int64     `json:"id"`
	IncidentID   int64     `json:"incidentId"`
	NoteText     string    `json:"noteText"`
	NoteSendDate time.Time `json:"noteSendDate"`
	NoteSendBy   *string   `json:"noteSendBy,omitempty"`
}
