package dto

// FeedsStatus is what an operator can be told about the ThreatWinds feed:
// whether this instance sends to it and whether it has credentials. The
// credentials themselves are never part of an answer.
type FeedsStatus struct {
	Enabled    bool `json:"enabled"`
}

type FeedsToggleRequest struct {
	Enabled bool `json:"enabled"`
}

