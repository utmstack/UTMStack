package dto

type CreateFilterGroupRequest struct {
	ID               *int64  `json:"id"`
	GroupName        string  `json:"groupName"`
	GroupDescription *string `json:"groupDescription"`
}

type UpdateFilterGroupRequest struct {
	ID               int64   `json:"id"`
	GroupName        string  `json:"groupName"`
	GroupDescription *string `json:"groupDescription"`
}

type FilterGroupResponse struct {
	ID               int64   `json:"id"`
	GroupName        string  `json:"groupName"`
	GroupDescription *string `json:"groupDescription"`
	SystemOwner      bool    `json:"systemOwner"`
}

type FilterGroupCountFilters struct {
	IDEquals                 *int64  `form:"id.equals"`
	GroupNameContains        *string `form:"groupName.contains"`
	GroupDescriptionContains *string `form:"groupDescription.contains"`
}

type FilterGroupListFilters struct {
	Page                     int     `form:"page"`
	Size                     int     `form:"size"`
	IDEquals                 *int64  `form:"id.equals"`
	GroupNameContains        *string `form:"groupName.contains"`
	GroupDescriptionContains *string `form:"groupDescription.contains"`
}
