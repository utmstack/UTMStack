package domain

import "time"

type UtmIncident struct {
	ID                  int64   `gorm:"primaryKey;autoIncrement"                           json:"id"`
	IncidentName        string  `gorm:"column:incident_name;size:255;not null;uniqueIndex"  json:"incidentName"`
	IncidentDescription *string `gorm:"column:incident_description;type:text"              json:"incidentDescription,omitempty"`
	IncidentStatus      string  `gorm:"column:incident_status;size:50;not null"            json:"incidentStatus"`
	IncidentSeverity    *int    `gorm:"column:incident_severity"                           json:"incidentSeverity,omitempty"`
	IncidentAssignedTo  *string `gorm:"column:incident_assigned_to;type:text"              json:"incidentAssignedTo,omitempty"`
	IncidentSolution    *string `gorm:"column:incident_solution;type:text"                 json:"incidentSolution,omitempty"`
	IncidentCreatedDate time.Time `gorm:"column:incident_created_date;not null;default:now()" json:"incidentCreatedDate"`
}

func (UtmIncident) TableName() string { return "utm_incident" }
