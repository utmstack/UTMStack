package dto

import "github.com/google/uuid"

type IncidentAlertRequest struct {
	IncidentID    uuid.UUID `json:"incidentId"    binding:"required"`
	AlertID       string    `json:"alertId"       binding:"required"`
	AlertName     string    `json:"alertName"     binding:"required"`
	AlertSeverity string    `json:"alertSeverity" binding:"required"`
	AlertStatus   *string   `json:"alertStatus"`
}

type UpdateAlertStatusRequest struct {
	IncidentID  uuid.UUID `json:"incidentId"  binding:"required"`
	AlertIds    []string  `json:"alertIds"    binding:"required,min=1"`
	AlertStatus string    `json:"status"      binding:"required"`
}

type UpdateIncidentAlertRequest struct {
	ID            uuid.UUID `json:"id"            binding:"required"`
	IncidentID    uuid.UUID `json:"incidentId"    binding:"required"`
	AlertID       string    `json:"alertId"       binding:"required"`
	AlertName     string    `json:"alertName"`
	AlertSeverity string    `json:"alertSeverity"`
	AlertStatus   *string   `json:"alertStatus"`
}
