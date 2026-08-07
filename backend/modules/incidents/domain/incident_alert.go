package domain

import "github.com/google/uuid"

type IncidentAlert struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID      uuid.UUID        `gorm:"column:tenant_id;type:uuid;not null;index:idx_incident_alert_tenant_incident,priority:1;uniqueIndex:uniq_incident_alert_tenant_alert,priority:1" json:"-"`
	IncidentID    uuid.UUID        `gorm:"column:incident_id;type:uuid;not null;index:idx_incident_alert_tenant_incident,priority:2"                                                      json:"incidentId"`
	AlertID       string           `gorm:"column:alert_id;size:255;not null;uniqueIndex:uniq_incident_alert_tenant_alert,priority:2" json:"alertId"`
	AlertName     string           `gorm:"column:alert_name;size:255;not null"                                                     json:"alertName"`
	AlertSeverity IncidentSeverity `gorm:"column:alert_severity;type:varchar(16);not null"                                         json:"alertSeverity"`
	AlertStatus   string           `gorm:"column:alert_status;type:varchar(32)"                                                    json:"alertStatus,omitempty"`
}

func (IncidentAlert) TableName() string { return "incident_alert" }
