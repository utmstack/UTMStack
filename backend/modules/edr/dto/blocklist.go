package dto

// AllowlistEntry is one allowlist row.
type AllowlistEntry struct {
	Category string `json:"category"` // allow_paths | allow_processes | allow_commands
	Value    string `json:"value"`
}

// BlocklistStatus is the blocklist feed health folded from statuses.
type BlocklistStatus struct {
	Indicators     int   `json:"indicators"`
	ActiveFilters  int   `json:"activeFilters"`
	FeedStale      bool  `json:"feedStale"`
	FeedAgeSec     int64 `json:"feedAgeSec"`
	RecentBlocks   int   `json:"recentBlocks"`
	Enforced       bool  `json:"enforced"`
	EnforcementDeg bool  `json:"enforcementDegraded"`
}
