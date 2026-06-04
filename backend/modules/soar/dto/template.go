package dto

type TemplateResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Command     string `json:"command"`
	SystemOwner bool   `json:"systemOwner"`
}

type TemplateFilters struct {
	// id.equals — exact match on template ID (JHipster: LongFilter)
	ID int64 `form:"id.equals"`
	// label.contains — substring match on label/title (JHipster: StringFilter)
	Label string `form:"label.contains"`
	// description.contains — substring match on description (JHipster: StringFilter)
	Description string `form:"description.contains"`
	// command.contains — substring match on command (JHipster: StringFilter)
	Command string `form:"command.contains"`
	// systemOwner.equals — filter by system ownership (JHipster: BooleanFilter)
	SystemOwner *bool `form:"systemOwner.equals"`
	// Page is 0-based (matches Java Spring Pageable). Default 0.
	Page int `form:"page,default=0"`
	// Size is the page size (matches Java). Default 20.
	Size int `form:"size,default=20"`
}
