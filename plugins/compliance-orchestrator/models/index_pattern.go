package models

type IndexPattern struct {
	ID            int    `json:"id"`
	Pattern       string `json:"pattern"`
	PatternModule string `json:"patternModule"`
	PatternSystem bool   `json:"patternSystem"`
	Active        bool   `json:"active"`
}
