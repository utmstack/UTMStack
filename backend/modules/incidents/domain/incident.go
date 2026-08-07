package domain

import (
	"time"

	"github.com/google/uuid"
)

type IncidentStatus string

const (
	StatusOpen      IncidentStatus = "Open"
	StatusInReview  IncidentStatus = "In review"
	StatusCompleted IncidentStatus = "Completed"
	StatusMerged    IncidentStatus = "Merged"
)

func (s IncidentStatus) Valid() bool {
	switch s {
	case StatusOpen, StatusInReview, StatusCompleted, StatusMerged:
		return true
	}
	return false
}

type IncidentSeverity string

const (
	SeverityLow    IncidentSeverity = "low"
	SeverityMedium IncidentSeverity = "medium"
	SeverityHigh   IncidentSeverity = "high"
)

func (s IncidentSeverity) Valid() bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh:
		return true
	}
	return false
}

func (s IncidentSeverity) Rank() int {
	switch s {
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	}
	return 0
}

type Incident struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"      json:"id"`
	TenantID    uuid.UUID        `gorm:"column:tenant_id;type:uuid;not null;index;uniqueIndex:uniq_incident_tenant_name" json:"-"`
	Name        string           `gorm:"column:incident_name;size:255;not null;uniqueIndex:uniq_incident_tenant_name" json:"incidentName"`
	Description *string          `gorm:"column:incident_description;type:text"              json:"incidentDescription,omitempty"`
	Status      IncidentStatus   `gorm:"column:incident_status;type:varchar(32);not null"   json:"incidentStatus"`
	Severity    IncidentSeverity `gorm:"column:incident_severity;type:varchar(16);not null;default:''" json:"incidentSeverity,omitempty"`
	AssignedTo  string           `gorm:"column:incident_assigned_to;type:text;not null;default:''"  json:"incidentAssignedTo,omitempty"`
	Solution    *string          `gorm:"column:incident_solution;type:text"                 json:"incidentSolution,omitempty"`
	CreatedDate time.Time        `gorm:"column:incident_created_date;not null;default:now()" json:"incidentCreatedDate"`

	// AlertCount is filled in on list reads; it is not a column. The list shows
	// how big each incident is, and loading every linked row to count them would
	// be the whole table to render one number per line.
	AlertCount int `gorm:"-" json:"alertCount"`

	Alerts  []IncidentAlert   `gorm:"foreignKey:IncidentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Notes   []IncidentNote    `gorm:"foreignKey:IncidentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	History []IncidentHistory `gorm:"foreignKey:IncidentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Incident) TableName() string { return "incident" }
