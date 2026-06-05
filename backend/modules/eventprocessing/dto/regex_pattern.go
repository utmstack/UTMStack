package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
)

type CreateRegexPatternRequest struct {
	ID                 *int64 `json:"id"`
	PatternID          string `json:"patternId"`
	PatternDescription string `json:"patternDescription"`
	PatternDefinition  string `json:"patternDefinition"`
}

type UpdateRegexPatternRequest struct {
	ID                 *int64 `json:"id"`
	PatternID          string `json:"patternId"`
	PatternDescription string `json:"patternDescription"`
	PatternDefinition  string `json:"patternDefinition"`
}

type RegexPatternResponse struct {
	ID                 int64      `json:"id"`
	PatternID          string     `json:"patternId"`
	PatternDescription string     `json:"patternDescription"`
	PatternDefinition  string     `json:"patternDefinition"`
	SystemOwner        bool       `json:"systemOwner"`
	LastUpdate         *time.Time `json:"lastUpdate"`
}

type RegexPatternFilters struct {
	// Page is 0-based (matches Java Spring Pageable).
	Page   int    `form:"page"`
	Size   int    `form:"size"`
	Search string `form:"search"`
}

func RegexPatternToResponse(e *domain.UtmRegexPattern) *RegexPatternResponse {
	return &RegexPatternResponse{
		ID:                 e.ID,
		PatternID:          e.PatternID,
		PatternDescription: e.PatternDescription,
		PatternDefinition:  e.PatternDefinition,
		SystemOwner:        e.SystemOwner,
		LastUpdate:         e.LastUpdate,
	}
}
