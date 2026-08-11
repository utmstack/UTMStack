package dto

// FeedsStatus is what an operator can be told about the ThreatWinds feed:
// whether this instance sends to it and whether it has credentials. The
// credentials themselves are never part of an answer.
type FeedsStatus struct {
	Enabled    bool `json:"enabled"`
	Configured bool `json:"configured"`
}

type FeedsToggleRequest struct {
	Enabled bool `json:"enabled"`
}

// FeedsCredentialsRequest is what the plugin sends after registering with
// ThreatWinds. It is stored encrypted and handed back only through the file the
// plugin reads.
type FeedsCredentialsRequest struct {
	APIKey    string `json:"apiKey" binding:"required"`
	APISecret string `json:"apiSecret" binding:"required"`
}
