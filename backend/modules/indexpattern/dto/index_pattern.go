package dto

type CreateIndexPatternRequest struct {
	ID            *int64  `json:"id"`
	Pattern       string  `json:"pattern"       binding:"required"`
	PatternModule *string `json:"patternModule"`
	PatternSystem *bool   `json:"patternSystem"`
	IsActive      *bool   `json:"isActive"`
}

type UpdateIndexPatternRequest struct {
	ID            int64   `json:"id"            binding:"required"`
	Pattern       string  `json:"pattern"       binding:"required"`
	PatternModule *string `json:"patternModule"`
	PatternSystem *bool   `json:"patternSystem"`
	IsActive      *bool   `json:"isActive"`
}

type IndexPatternResponse struct {
	ID            int64   `json:"id"`
	Pattern       string  `json:"pattern"`
	PatternModule *string `json:"patternModule"`
	PatternSystem *bool   `json:"patternSystem"`
	IsActive      *bool   `json:"isActive"`
}

type IndexPatternFilters struct {
	IDEquals        *int64  `form:"id.equals"`
	IDGreaterThan   *int64  `form:"id.greaterThan"`
	IDLessThan      *int64  `form:"id.lessThan"`
	PatternContains *string `form:"pattern.contains"`
	PatternEquals   *string `form:"pattern.equals"`
	ModuleContains  *string `form:"patternModule.contains"`
	ModuleEquals    *string `form:"patternModule.equals"`
	PatternSystem   *bool   `form:"patternSystem.equals"`
	IsActive        *bool   `form:"isActive.equals"`
	Page            int     `form:"page"`
	Size            int     `form:"size"`
	Sort            string  `form:"sort"`
}

type IndexField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type IndexPatternFieldsResponse struct {
	IndexPattern IndexPatternResponse `json:"indexPattern"`
	Fields       []IndexField         `json:"fields"`
}
