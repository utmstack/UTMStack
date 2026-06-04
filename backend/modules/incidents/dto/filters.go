package dto

import "time"

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
	IncidentID  *int64
	AlertID     *string
	AlertStatus *int
	Page        int
	Size        int
	Sort        string
}

type IncidentNoteListQuery struct {
	IncidentID *int64
	Page       int
	Size       int
	Sort       string
}
