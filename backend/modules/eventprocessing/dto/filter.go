package dto

// CreateFilterRequest creates a new user filter. RelPath is the relative path
// under the user overlay (e.g. "myorg/custom-fw.yaml"). Content must be a
// valid pipeline: YAML as the go-sdk Config proto expects.
type CreateFilterRequest struct {
	RelPath string `json:"relPath" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// UpdateFilterRequest replaces the content of an existing user filter.
type UpdateFilterRequest struct {
	RelPath string `json:"relPath" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// FilterResponse is the API representation of a filter overlay entry.
type FilterResponse struct {
	RelPath string `json:"relPath"`
	Content string `json:"content"`
	System  bool   `json:"system"`
	Active  bool   `json:"active"`
}

// FilterFilters are query parameters for the list endpoint.
type FilterFilters struct {
	RelPathContains *string `form:"relPath.contains"`
	IsActiveEq      *bool   `form:"isActive.equals"`
	SystemEq        *bool   `form:"system.equals"`
	Page            int     `form:"page"`
	Size            int     `form:"size"`
}
