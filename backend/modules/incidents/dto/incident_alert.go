package dto

type IncidentAlertRequest struct {
	IncidentID    int64  `json:"incidentId"    binding:"required"`
	AlertID       string `json:"alertId"       binding:"required"`
	AlertName     string `json:"alertName"     binding:"required"`
	AlertSeverity int    `json:"alertSeverity" binding:"required"`
	AlertStatus   *int   `json:"alertStatus"`
}

type UpdateAlertStatusRequest struct {
	IncidentID  int64    `json:"incidentId"  binding:"required"`
	AlertIds    []string `json:"alertIds"    binding:"required,min=1"`
	AlertStatus int      `json:"status"      binding:"required"`
}

type UpdateIncidentAlertRequest struct {
	ID            int64  `json:"id"            binding:"required"`
	IncidentID    int64  `json:"incidentId"    binding:"required"`
	AlertID       string `json:"alertId"       binding:"required"`
	AlertName     string `json:"alertName"`
	AlertSeverity int    `json:"alertSeverity"`
	AlertStatus   *int   `json:"alertStatus"`
}

type IncidentAlertResponse struct {
	ID            int64  `json:"id"`
	IncidentID    int64  `json:"incidentId"`
	AlertID       string `json:"alertId"`
	AlertName     string `json:"alertName"`
	AlertSeverity int    `json:"alertSeverity"`
	AlertStatus   *int   `json:"alertStatus,omitempty"`
}
