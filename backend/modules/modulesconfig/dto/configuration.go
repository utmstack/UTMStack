package dto

import "github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"

type GroupConfigurationItem struct {
	ID              int64  `json:"id"`
	GroupID         int64  `json:"groupId"`
	ConfKey         string `json:"confKey"`
	ConfValue       string `json:"confValue"`
	ConfName        string `json:"confName"`
	ConfDescription string `json:"confDescription"`
	ConfDataType    string `json:"confDataType"`
	ConfRequired    bool   `json:"confRequired"`
	ConfOptions     string `json:"confOptions,omitempty"`
	ConfVisibility  string `json:"confVisibility,omitempty"`
}

func FromConfiguration(c domain.UtmModuleGroupConfiguration, reveal bool) GroupConfigurationItem {
	item := GroupConfigurationItem{
		ID:              c.ID,
		GroupID:         c.GroupID,
		ConfKey:         c.ConfKey,
		ConfValue:       c.ConfValue,
		ConfName:        c.ConfName,
		ConfDescription: c.ConfDescription,
		ConfDataType:    c.ConfDataType,
		ConfRequired:    c.ConfRequired,
		ConfOptions:     c.ConfOptions,
		ConfVisibility:  c.ConfVisibility,
	}
	if !reveal && domain.IsSensitive(c.ConfDataType) && c.ConfValue != "" {
		item.ConfValue = domain.MaskedValue
	}
	return item
}

type UpdateGroupConfigurationRequest struct {
	ModuleID int64                          `json:"moduleId" binding:"required"`
	Keys     []GroupConfigurationItemInput `json:"keys" binding:"required,dive"`
}

type GroupConfigurationItemInput struct {
	ID              int64  `json:"id"`
	GroupID         int64  `json:"groupId" binding:"required"`
	ConfKey         string `json:"confKey" binding:"required"`
	ConfValue       string `json:"confValue"`
	ConfName        string `json:"confName"`
	ConfDescription string `json:"confDescription"`
	ConfDataType    string `json:"confDataType" binding:"required"`
	ConfRequired    bool   `json:"confRequired"`
	ConfOptions     string `json:"confOptions"`
	ConfVisibility  string `json:"confVisibility"`
}

func (i GroupConfigurationItemInput) ToConfiguration() domain.UtmModuleGroupConfiguration {
	return domain.UtmModuleGroupConfiguration{
		ID:              i.ID,
		GroupID:         i.GroupID,
		ConfKey:         i.ConfKey,
		ConfValue:       i.ConfValue,
		ConfName:        i.ConfName,
		ConfDescription: i.ConfDescription,
		ConfDataType:    i.ConfDataType,
		ConfRequired:    i.ConfRequired,
		ConfOptions:     i.ConfOptions,
		ConfVisibility:  i.ConfVisibility,
	}
}
