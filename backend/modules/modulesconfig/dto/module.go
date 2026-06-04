package dto

import "github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"

// ModuleActivationRequest mirrors the legacy /utm-modules/activateDeactivate
// query params: which server, which module-kind, and whether to enable it.
type ModuleActivationRequest struct {
	ServerID         int64  `form:"serverId" binding:"required"`
	ModuleName       string `form:"nameShort" binding:"required"`
	ActivationStatus *bool  `form:"activationStatus" binding:"required"`
}

// ModuleResponse is the outward shape of a UtmModule row, with its groups +
// configurations expanded (masked when not requested decrypted).
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

// FromModule builds a ModuleResponse with groups/configurations expanded.
// reveal=false masks sensitive values; reveal=true returns them verbatim and is
// only used by the internal-key-gated decrypted endpoint.
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
