package dto

type CreatePipelineRequest struct {
	RelPath string `json:"relPath" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type UpdatePipelineRequest struct {
	RelPath string `json:"relPath" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// Order is the full sequence of pipeline names this tenant wants, not one
// position: it is stored as a list and a partial update could not be resolved.
type UpdatePipelineOrderRequest struct {
	Order []string `json:"order"`
}

type PipelineResponse struct {
	RelPath   string   `json:"relPath"`
	Content   string   `json:"content"`
	System    bool     `json:"system"`
	Active    bool     `json:"active"`
	DataTypes []string `json:"dataTypes"`
	Order     int32    `json:"order"`
}

type PipelineFilters struct {
	RelPathContains *string `form:"relPath.contains"`
	IsActiveEq      *bool   `form:"isActive.equals"`
	SystemEq        *bool   `form:"system.equals"`
	DataTypeEq      *string `form:"dataType.equals"`
	Page            int     `form:"page"`
	Size            int     `form:"size"`
}
