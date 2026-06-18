package dto

// IngestionStatsBucket is one (dataSource|dataType) total over the window.
// LastSeen is the most recent @timestamp of a statistics doc for the key within
// the window (RFC3339) — i.e. the last time logs were observed for that source.
type IngestionStatsBucket struct {
	Key      string `json:"key"`
	Count    int64  `json:"count"`
	LastSeen string `json:"lastSeen,omitempty"`
}

// IngestionStatsResponse is the totals-by-bucket view (read live from
// v11-statistics-*; no Postgres/sync).
type IngestionStatsResponse struct {
	GroupBy string                 `json:"groupBy"`
	Status  string                 `json:"status"`
	From    string                 `json:"from"`
	To      string                 `json:"to"`
	Total   int64                  `json:"total"`
	Buckets []IngestionStatsBucket `json:"buckets"`
}

// TimelinePoint is one event count at a time bucket.
type TimelinePoint struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
}

// TimelineSeries is a per-key line (only when groupBy is set).
type TimelineSeries struct {
	Key    string          `json:"key"`
	Points []TimelinePoint `json:"points"`
}

// IngestionTimelineResponse is the time-series view. Without groupBy, Points is
// populated; with groupBy, Series carries one line per dataSource/dataType.
type IngestionTimelineResponse struct {
	Status   string           `json:"status"`
	GroupBy  string           `json:"groupBy,omitempty"`
	Interval string           `json:"interval"`
	From     string           `json:"from"`
	To       string           `json:"to"`
	Points   []TimelinePoint  `json:"points,omitempty"`
	Series   []TimelineSeries `json:"series,omitempty"`
}
