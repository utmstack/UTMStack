package dto

import "github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"

// CreateModuleGroupRequest is the POST body for /utm-configuration-groups.
// The legacy panel sends moduleId + a human-readable name/description, plus an
// optional collector id when the group belongs to a specific agent collector.
type CreateModuleGroupRequest struct {
	ModuleID         int64   `json:"moduleId" binding:"required"`
	GroupName        string  `json:"groupName" binding:"required"`
	GroupDescription string  `json:"groupDescription"`
	Collector        *string `json:"collector"`
}

// UpdateModuleGroupRequest is the PUT body. The legacy controller required the
// id in the payload; we keep that shape so the panel keeps working unchanged.
type UpdateModuleGroupRequest struct {
	ID               int64   `json:"id" binding:"required"`
	GroupName        string  `json:"groupName"`
	GroupDescription string  `json:"groupDescription"`
	Collector        *string `json:"collector"`
}

// ModuleGroupResponse is the outward shape, including the expanded config rows.
type ModuleGroupResponse struct {
	ID                        int64                    `json:"id"`
	ModuleID                  int64                    `json:"moduleId"`
	GroupName                 string                   `json:"groupName"`
	GroupDescription          string                   `json:"groupDescription,omitempty"`
	Collector                 *string                  `json:"collector,omitempty"`
	ModuleGroupConfigurations []GroupConfigurationItem `json:"moduleGroupConfigurations"`
}

// FromGroup builds a ModuleGroupResponse. reveal=false masks sensitive values.
func FromGroup(g domain.UtmModuleGroup, reveal bool) ModuleGroupResponse {
	out := ModuleGroupResponse{
		ID:                        g.ID,
		ModuleID:                  g.ModuleID,
		GroupName:                 g.GroupName,
		GroupDescription:          g.GroupDescription,
		Collector:                 g.Collector,
		ModuleGroupConfigurations: make([]GroupConfigurationItem, 0, len(g.ModuleGroupConfigurations)),
	}
	for _, c := range g.ModuleGroupConfigurations {
		out.ModuleGroupConfigurations = append(out.ModuleGroupConfigurations, FromConfiguration(c, reveal))
	}
	return out
}
