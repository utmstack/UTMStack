package dto

// EnabledRequest is the body for the enable/disable toggle endpoints.
type EnabledRequest struct {
	Enabled bool `json:"enabled"`
}
