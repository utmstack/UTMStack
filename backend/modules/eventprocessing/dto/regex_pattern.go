package dto

// Identity for regex patterns is patternId (the {{.name}} reference used in
// filter YAMLs). System patterns cannot be modified or deleted.

type CreateRegexPatternRequest struct {
	PatternID          string `json:"patternId"          binding:"required"`
	PatternDescription string `json:"patternDescription"`
	PatternDefinition  string `json:"patternDefinition"  binding:"required"`
}

type UpdateRegexPatternRequest struct {
	PatternID          string `json:"patternId"          binding:"required"`
	PatternDescription string `json:"patternDescription"`
	PatternDefinition  string `json:"patternDefinition"  binding:"required"`
}

type RegexPatternResponse struct {
	PatternID          string `json:"patternId"`
	PatternDescription string `json:"patternDescription"`
	PatternDefinition  string `json:"patternDefinition"`
	SystemOwner        bool   `json:"systemOwner"`
}

type RegexPatternFilters struct {
	Search string `form:"search"`
	System *bool  `form:"system"` // nil = both; true = system only; false = user only
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}
