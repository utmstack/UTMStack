package dto

import "github.com/utmstack/utmstack/backend/modules/alerts/domain"

type UpdateAlertStatusRequest struct {
	AlertIDs            []string           `json:"alertIds" binding:"required"`
	Status              domain.AlertStatus `json:"status" binding:"required"`
	StatusObservation   string             `json:"statusObservation"`
	AddFalsePositiveTag bool               `json:"addFalsePositiveTag"`
}

type UpdateAlertTagsRequest struct {
	AlertIDs   []string `json:"alertIds" binding:"required"`
	Tags       []string `json:"tags"`
	CreateRule bool     `json:"createRule"`
}

type UpdateAlertAssigneeRequest struct {
	AlertID  string `json:"alertId" binding:"required"`
	Assignee string `json:"assignee"` // empty clears the assignment
}

type ConvertToIncidentRequest struct {
	AlertIDs       []string `json:"eventIds"       binding:"required"`
	IncidentName   string   `json:"incidentName"   binding:"required"`
	IncidentID     string   `json:"incidentId"     binding:"required"`
	IncidentSource string   `json:"incidentSource"`
}

type CountOpenAlertsResponse struct {
	Count int64 `json:"count"`
}

type NotifyAlertRequest struct {
	AlertID string `json:"alertId" binding:"required"`
}
