package dto

// Identity for regex patterns is patternId (the {{.name}} reference used in
// filter YAMLs). Patterns are read-only — a shared vocabulary seeded by the
// pipeline bootstrap — so there are no create or update request types.

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
