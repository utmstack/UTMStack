package dto

import "github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"

type ModuleActivationRequest struct {
	ServerID         int64  `form:"serverId" binding:"required"`
	ModuleName       string `form:"nameShort" binding:"required"`
	ActivationStatus *bool  `form:"activationStatus" binding:"required"`
}

type ModuleResponse struct {
	ID                int64                 `json:"id"`
	ServerID          int64                 `json:"serverId"`
	PrettyName        string                `json:"prettyName,omitempty"`
	ModuleName        string                `json:"moduleName"`
	ModuleDescription string                `json:"moduleDescription,omitempty"`
	ModuleActive      bool                  `json:"moduleActive"`
	ModuleIcon        string                `json:"moduleIcon,omitempty"`
	ModuleCategory    string                `json:"moduleCategory,omitempty"`
	LiteVersion       bool                  `json:"liteVersion"`
	NeedsRestart      bool                  `json:"needsRestart"`
	IsActivatable     bool                  `json:"isActivatable"`
	ModuleGroups      []ModuleGroupResponse `json:"moduleGroups"`
}

func FromModule(m domain.UtmModule, reveal bool) ModuleResponse {
	resp := ModuleResponse{
		ID:                m.ID,
		ServerID:          m.ServerID,
		PrettyName:        m.PrettyName,
		ModuleName:        m.ModuleName,
		ModuleDescription: m.ModuleDescription,
		ModuleActive:      m.ModuleActive,
		ModuleIcon:        m.ModuleIcon,
		ModuleCategory:    m.ModuleCategory,
		LiteVersion:       m.LiteVersion,
		NeedsRestart:      m.NeedsRestart,
		IsActivatable:     m.IsActivatable,
		ModuleGroups:      make([]ModuleGroupResponse, 0, len(m.ModuleGroups)),
	}
	for _, g := range m.ModuleGroups {
		resp.ModuleGroups = append(resp.ModuleGroups, FromGroup(g, reveal))
	}
	return resp
}
