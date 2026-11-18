package dto

// OverviewResponse is the fleet summary.
type OverviewResponse struct {
	Endpoints       int   `json:"endpoints"`
	WithModule      int   `json:"withModule"`
	Healthy         int   `json:"healthy"`
	Unhealthy       int   `json:"unhealthy"`
	StaleSignatures int   `json:"staleSignatures"`
	Drifted         int   `json:"drifted"`
	DetectionsToday int64 `json:"detectionsToday"`
}
