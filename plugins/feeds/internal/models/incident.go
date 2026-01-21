package models

import "time"

type Incident struct {
	ID          int64     `json:"id"`
	Name        string    `json:"incidentName"`
	Description string    `json:"incidentDescription"`
	Status      string    `json:"incidentStatus"`
	Severity    int       `json:"incidentSeverity"`
	CreatedDate time.Time `json:"incidentCreatedDate"`
}

type IncidentAlert struct {
	ID            int64  `json:"id"`
	IncidentID    int64  `json:"incidentId"`
	AlertID       string `json:"alertId"`
	AlertName     string `json:"alertName"`
	AlertStatus   int    `json:"alertStatus"`
	AlertSeverity int    `json:"alertSeverity"`
}
