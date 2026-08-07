package domain

import (
	"time"

	"github.com/google/uuid"
)

type UtmIncident struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement"                           json:"id"`
	TenantID            uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index;uniqueIndex:idx_incident_tenant_name" json:"-"`
	IncidentName        string    `gorm:"column:incident_name;size:255;not null;uniqueIndex:idx_incident_tenant_name" json:"incidentName"`
	IncidentDescription *string   `gorm:"column:incident_description;type:text"              json:"incidentDescription,omitempty"`
	IncidentStatus      string    `gorm:"column:incident_status;size:255;not null"           json:"incidentStatus"`
	IncidentSeverity    *int      `gorm:"column:incident_severity"                           json:"incidentSeverity,omitempty"`
	IncidentAssignedTo  *string   `gorm:"column:incident_assigned_to;type:text"              json:"incidentAssignedTo,omitempty"`
	IncidentSolution    *string   `gorm:"column:incident_solution;type:text"                 json:"incidentSolution,omitempty"`
	IncidentCreatedDate time.Time `gorm:"column:incident_created_date;not null;default:now()" json:"incidentCreatedDate"`

	Alerts  []UtmIncidentAlert   `gorm:"foreignKey:IncidentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Notes   []UtmIncidentNote    `gorm:"foreignKey:IncidentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	History []UtmIncidentHistory `gorm:"foreignKey:IncidentID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (UtmIncident) TableName() string { return "incident" }
