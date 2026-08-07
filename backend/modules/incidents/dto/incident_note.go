package dto

import "github.com/google/uuid"

type CreateNoteRequest struct {
	IncidentID uuid.UUID `json:"incidentId" binding:"required"`
	NoteText   string    `json:"noteText"   binding:"required,max=1000"`
}
