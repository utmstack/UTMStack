package dto

import "time"

type CreateFilterRequest struct {
	ID             *int64  `json:"id"`
	FilterName     string  `json:"filterName"`
	LogstashFilter string  `json:"logstashFilter"`
	FilterGroupID  *int64  `json:"filterGroupId"`
	DataTypeID     *int64  `json:"dataTypeId"`
	SystemOwner    bool    `json:"systemOwner"`
	IsActive       bool    `json:"isActive"`
	ModuleName     *string `json:"moduleName"`
	FilterVersion  *string `json:"filterVersion"`
}

type UpdateFilterRequest struct {
	ID             int64   `json:"id"`
	FilterName     string  `json:"filterName"`
	LogstashFilter string  `json:"logstashFilter"`
	FilterGroupID  *int64  `json:"filterGroupId"`
	DataTypeID     *int64  `json:"dataTypeId"`
	SystemOwner    bool    `json:"systemOwner"`
	IsActive       bool    `json:"isActive"`
	ModuleName     *string `json:"moduleName"`
	FilterVersion  *string `json:"filterVersion"`
}

type FilterResponse struct {
	ID             int64      `json:"id"`
	FilterName     string     `json:"filterName"`
	LogstashFilter string     `json:"logstashFilter"`
	FilterGroupID  *int64     `json:"filterGroupId"`
	SystemOwner    bool       `json:"systemOwner"`
	IsActive       bool       `json:"isActive"`
	ModuleName     *string    `json:"moduleName"`
	FilterVersion  *string    `json:"filterVersion"`
	UpdatedAt      *time.Time `json:"updatedAt"`
}

type FilterFilters struct {
	IDEquals           *int64  `form:"id.equals"`
	FilterNameContains *string `form:"filterName.contains"`
	FilterGroupIDEq    *int64  `form:"filterGroupId.equals"`
	FilterGroupIDGte   *int64  `form:"filterGroupId.greaterThanOrEqual"`
	FilterGroupIDLte   *int64  `form:"filterGroupId.lessThanOrEqual"`
	IsActiveEq         *bool   `form:"isActive.equals"`
	Page               int     `form:"page"`
	Size               int     `form:"size"`
}
