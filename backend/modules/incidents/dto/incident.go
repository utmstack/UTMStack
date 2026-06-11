package dto

import "time"

type AlertLinkItem struct {
	AlertID       string `json:"alertId"       binding:"required"`
	AlertName     string `json:"alertName"     binding:"required"`
	AlertStatus   *int   `json:"alertStatus"`
	AlertSeverity int    `json:"alertSeverity" binding:"required"`
}

type CreateIncidentRequest struct {
	IncidentName        string          `json:"incidentName"        binding:"required"`
	IncidentDescription *string         `json:"incidentDescription"`
	IncidentAssignedTo  *string         `json:"incidentAssignedTo"`
	IncidentObservation *string         `json:"incidentObservation"`
	AlertList           []AlertLinkItem `json:"alertList"           binding:"required,min=1"`
}

type AddAlertsRequest struct {
	IncidentID int64           `json:"incidentId" binding:"required"`
	AlertList  []AlertLinkItem `json:"alertList"  binding:"required,min=1"`
}

type ChangeStatusRequest struct {
	IncidentID          int64   `json:"id"                  binding:"required"`
	IncidentName        string  `json:"incidentName"        binding:"required"`
	IncidentDescription *string `json:"incidentDescription"`
	IncidentStatus      string  `json:"incidentStatus"      binding:"required"`
	IncidentSeverity    *int    `json:"incidentSeverity"`
	IncidentCreatedDate *string `json:"incidentCreatedDate"`
	IncidentSolution    *string `json:"incidentSolution"`
}

type AssignRequest struct {
	IncidentID int64   `json:"incidentId" binding:"required"`
	AssignedTo *string `json:"assignedTo"`
}

type IncidentResponse struct {
	ID                  int64     `json:"id"`
	IncidentName        string    `json:"incidentName"`
	IncidentDescription *string   `json:"incidentDescription,omitempty"`
	IncidentStatus      string    `json:"incidentStatus"`
	IncidentSeverity    *int      `json:"incidentSeverity,omitempty"`
	IncidentAssignedTo  *string   `json:"incidentAssignedTo,omitempty"`
	IncidentSolution    *string   `json:"incidentSolution,omitempty"`
	IncidentCreatedDate time.Time `json:"incidentCreatedDate"`
	AlertCount          int       `json:"alertCount"`
}
