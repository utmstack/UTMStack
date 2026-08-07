package dto

import "github.com/google/uuid"

type AlertLinkItem struct {
	AlertID       string  `json:"alertId"       binding:"required"`
	AlertName     string  `json:"alertName"     binding:"required"`
	AlertStatus   *string `json:"alertStatus"`
	AlertSeverity string  `json:"alertSeverity" binding:"required"`
}

type CreateIncidentRequest struct {
	IncidentName        string  `json:"incidentName"        binding:"required"`
	IncidentDescription *string `json:"incidentDescription"`
	IncidentAssignedTo  string  `json:"incidentAssignedTo"`
	IncidentObservation *string `json:"incidentObservation"`
	// Capped because nothing else caps it: the rows are written in one
	// transaction, and an unbounded list is a request that holds a Postgres
	// transaction open for as long as the caller likes.
	AlertList []AlertLinkItem `json:"alertList" binding:"required,min=1,max=1000"`
}

type AddAlertsRequest struct {
	IncidentID uuid.UUID       `json:"incidentId" binding:"required"`
	AlertList  []AlertLinkItem `json:"alertList"  binding:"required,min=1,max=1000"`
}

type ChangeStatusRequest struct {
	IncidentID          uuid.UUID `json:"id"                  binding:"required"`
	IncidentName        string    `json:"incidentName"        binding:"required"`
	IncidentDescription *string   `json:"incidentDescription"`
	IncidentStatus      string    `json:"incidentStatus"      binding:"required"`
	IncidentCreatedDate *string   `json:"incidentCreatedDate"`
	IncidentSolution    *string   `json:"incidentSolution"`
}

type AssignRequest struct {
	IncidentID uuid.UUID `json:"incidentId" binding:"required"`
	AssignedTo string    `json:"assignedTo"`
}
