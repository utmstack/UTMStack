package dto

import "github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"

type CreateModuleGroupRequest struct {
	ModuleID         int64   `json:"moduleId" binding:"required"`
	GroupName        string  `json:"groupName" binding:"required"`
	GroupDescription string  `json:"groupDescription"`
	Collector        *string `json:"collector"`
}

type UpdateModuleGroupRequest struct {
	ID               int64   `json:"id" binding:"required"`
	GroupName        string  `json:"groupName"`
	GroupDescription string  `json:"groupDescription"`
	Collector        *string `json:"collector"`
}

type ModuleGroupResponse struct {
	ID                        int64                    `json:"id"`
	ModuleID                  int64                    `json:"moduleId"`
	GroupName                 string                   `json:"groupName"`
	GroupDescription          string                   `json:"groupDescription,omitempty"`
	Collector                 *string                  `json:"collector,omitempty"`
	ModuleGroupConfigurations []GroupConfigurationItem `json:"moduleGroupConfigurations"`
}

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
