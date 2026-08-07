package dto

import "github.com/google/uuid"

type HistoryListQuery struct {
	IncidentID *uuid.UUID
	Action     *string
	Page       int
	Size       int
	Sort       string
}
