package dto

import (
	"time"

	"github.com/google/uuid"
)

type IncidentListQuery struct {
	IncidentName       *string
	IncidentStatus     *string
	IncidentAssignedTo *string
	CreatedDateStart   *time.Time
	CreatedDateEnd     *time.Time
	Page               int
	Size               int
	Sort               string
}

type IncidentAlertListQuery struct {
	IncidentID  *uuid.UUID
	AlertID     *string
	AlertStatus *string
	Page        int
	Size        int
	Sort        string
}

type IncidentNoteListQuery struct {
	IncidentID *uuid.UUID
	Page       int
	Size       int
	Sort       string
}
