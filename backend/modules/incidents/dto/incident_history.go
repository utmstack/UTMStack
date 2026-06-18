package dto

import "time"

type HistoryResponse struct {
	ID                int64     `json:"id"`
	IncidentID        int64     `json:"incidentId"`
	Action            string    `json:"action"`
	ActionType        string    `json:"actionType"`
	ActionDetail      *string   `json:"actionDetail,omitempty"`
	ActionCreatedDate time.Time `json:"actionCreatedDate"`
	ActionCreatedBy   *string   `json:"actionCreatedBy,omitempty"`
}

type HistoryListQuery struct {
	IncidentID *int64
	ActionType *string
	Page       int
	Size       int
	Sort       string
}
