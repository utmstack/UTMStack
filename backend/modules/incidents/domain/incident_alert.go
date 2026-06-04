package domain

type UtmIncidentAlert struct {
	ID            int64  `gorm:"primaryKey;autoIncrement"                   json:"id"`
	IncidentID    int64  `gorm:"column:incident_id;not null"                json:"incidentId"`
	AlertID       string `gorm:"column:alert_id;size:255;not null;uniqueIndex" json:"alertId"`
	AlertName     string `gorm:"column:alert_name;size:255;not null"         json:"alertName"`
	AlertSeverity int    `gorm:"column:alert_severity;not null"              json:"alertSeverity"`
	AlertStatus   *int   `gorm:"column:alert_status"                         json:"alertStatus,omitempty"`
}

func (UtmIncidentAlert) TableName() string { return "utm_incident_alert" }
