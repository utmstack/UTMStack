package dto

import "github.com/utmstack/utmstack/backend/modules/integrations/domain"

type ModuleActivationRequest struct {
	ModuleName       string `form:"moduleName" json:"moduleName" binding:"required"`
	ActivationStatus *bool  `form:"activationStatus" json:"activationStatus" binding:"required"`
}

type CreateModuleRequest struct {
	ModuleName        string `json:"moduleName" binding:"required"`
	DataType          string `json:"dataType" binding:"required"`
	PrettyName        string `json:"prettyName"`
	ModuleDescription string `json:"moduleDescription"`
	ModuleIcon        string `json:"moduleIcon"`
	ModuleCategory    string `json:"moduleCategory"`
}

type UpdateModuleRequest struct {
	PrettyName        string `json:"prettyName"`
	ModuleDescription string `json:"moduleDescription"`
	ModuleIcon        string `json:"moduleIcon"`
	ModuleCategory    string `json:"moduleCategory"`
}

type ModuleResponse struct {
	ID                int64  `json:"id"`
	DataType          string `json:"dataType,omitempty"`
	ModuleName        string `json:"moduleName"`
	PrettyName        string `json:"prettyName,omitempty"`
	ModuleDescription string `json:"moduleDescription,omitempty"`
	ModuleActive      bool   `json:"moduleActive"`
	ModuleIcon        string `json:"moduleIcon,omitempty"`
	ModuleCategory    string `json:"moduleCategory,omitempty"`
	IsSystem          bool   `json:"isSystem"`
}

type ModuleListFilter struct {
	ModuleCategory *string
	ModuleActive   *bool
	NameContains   *string
	Page           int
	Size           int
}

func FromModule(m domain.UtmModule) ModuleResponse {
	return ModuleResponse{
		ID:                m.ID,
		DataType:          m.DataType,
		ModuleName:        m.ModuleName,
		PrettyName:        m.PrettyName,
		ModuleDescription: m.ModuleDescription,
		ModuleActive:      m.ModuleActive,
		ModuleIcon:        m.ModuleIcon,
		ModuleCategory:    m.ModuleCategory,
		IsSystem:          m.IsSystem,
	}
}
