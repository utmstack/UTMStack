package dto

// FeedsStatus is what an operator can be told about the ThreatWinds feed:
// whether this instance sends to it and whether it has credentials. The
// credentials themselves are never part of an answer.
type FeedsStatus struct {
	FeedsEnabled    			bool `json:"feeds_enabled"`
	FeedsContributionEnabled    bool `json:"feeds_contribution_enabled"`
}

type FeedsToggleRequest struct {
	FeedsEnabled    			bool `json:"feeds_enabled"`
	FeedsContributionEnabled    bool `json:"feeds_contribution_enabled"`
}

