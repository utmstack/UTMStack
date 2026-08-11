package dto

type RegexPatternResponse struct {
	PatternID         string `json:"patternId"`
	PatternDefinition string `json:"patternDefinition"`
}

type RegexPatternFilters struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}
