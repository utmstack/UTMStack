package dto

import "github.com/utmstack/utmstack/backend/pkg/common_models"

// BulkSMTPField is one key→value pair to upsert.
type BulkSMTPField struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

// BulkSMTPUpdateRequest sets SMTP fields across N tenants.
type BulkSMTPUpdateRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	Fields   []BulkSMTPField                  `json:"fields" binding:"required,min=1"`
}

// BulkSMTPTestRequest sends a test mail from N tenants.
type BulkSMTPTestRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
}
