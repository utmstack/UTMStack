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
	RelPath   string   `json:"relPath"`
	Content   string   `json:"content"`
	System    bool     `json:"system"`
	Active    bool     `json:"active"`
	DataTypes []string `json:"dataTypes"`
	Order     int32    `json:"order"`
}

type UpdateFilterOrderRequest struct {
	RelPath string `json:"relPath" binding:"required"`
	Order   int32  `json:"order" binding:"required"`
}

// FilterFilters are query parameters for the list endpoint.
type FilterFilters struct {
	RelPathContains *string `form:"relPath.contains"`
	IsActiveEq      *bool   `form:"isActive.equals"`
	SystemEq        *bool   `form:"system.equals"`
	DataTypeEq      *string `form:"dataType.equals"`
	Page            int     `form:"page"`
	Size            int     `form:"size"`
}
