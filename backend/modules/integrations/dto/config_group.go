package dto

import "github.com/utmstack/utmstack/backend/modules/integrations/domain"

type ConfigGroupRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Config      map[string]string `json:"config"`
}

type ConfigGroupResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Config      map[string]string `json:"config"`
}

func FromConfigGroup(g domain.ConfigGroup) ConfigGroupResponse {
	return ConfigGroupResponse{Name: g.Name, Description: g.Description, Config: g.Config}
}
