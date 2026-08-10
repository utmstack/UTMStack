package dto

import (
	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
)

type CreateIntegrationRequest struct {
	Name        string            `json:"name"        binding:"required"`
	DataType    string            `json:"dataType"    binding:"required"`
	IngestType  domain.IngestType `json:"ingestType"  binding:"required"`
	Description string            `json:"description"`
	Icon        string            `json:"icon"`
}

type UpdateIntegrationRequest struct {
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type IntegrationResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	DataType    string            `json:"dataType,omitempty"`
	IngestType  domain.IngestType `json:"ingestType,omitempty"`
	Description string            `json:"description,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	SystemOwner bool              `json:"systemOwner"`

	// Configured reports whether this tenant has any configuration group for
	// the integration. It replaces the stored active flag, which said only
	// that somebody had pressed a switch.
	Configured bool `json:"configured"`
}

type DataTypeOption struct {
	DataType    string `json:"dataType"`
	Name        string `json:"name"`
	SystemOwner bool   `json:"systemOwner"`
}

func FromIntegration(i domain.Integration) IntegrationResponse {
	return IntegrationResponse{
		ID:          i.ID,
		Name:        i.Name,
		DataType:    i.DataType,
		IngestType:  i.IngestType,
		Description: i.Description,
		Icon:        i.Icon,
		SystemOwner: i.SystemOwner,
	}
}
