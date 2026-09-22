package models

import "time"

type Incident struct {
	ID          string    `json:"id"`
	Name        string    `json:"incidentName"`
	Description *string   `json:"incidentDescription"`
	Status      string    `json:"incidentStatus"`
	Severity    string    `json:"incidentSeverity"`
	CreatedDate time.Time `json:"incidentCreatedDate"`
	AlertCount  int       `json:"alertCount"`
}

type IncidentAlert struct {
	ID            string `json:"id"`
	IncidentID    string `json:"incidentId"`
	AlertID       string `json:"alertId"`
	AlertName     string `json:"alertName"`
	AlertSeverity string `json:"alertSeverity"`
	AlertStatus   string `json:"alertStatus"`
}
