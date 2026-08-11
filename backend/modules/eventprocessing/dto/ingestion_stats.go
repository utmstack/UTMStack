package dto

type IngestionStatsBucket struct {
	Key      string `json:"key"`
	Count    int64  `json:"count"`
	Bytes    int64  `json:"bytes"`
	LastSeen string `json:"lastSeen,omitempty"`
}

// IngestionTotals is the whole window: how many events arrived and how much
// they weighed. The volume is what a licence is measured against, so it is
// answered next to the count rather than derived from it.
type IngestionTotals struct {
	Events int64 `json:"events"`
	Bytes  int64 `json:"bytes"`
}

type IngestionStatsResponse struct {
	GroupBy    string                 `json:"groupBy"`
	Status     string                 `json:"status"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Total      int64                  `json:"total"`
	TotalBytes int64                  `json:"totalBytes"`
	Buckets    []IngestionStatsBucket `json:"buckets"`
}

type TimelinePoint struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
	Bytes     int64  `json:"bytes"`
}

type TimelineSeries struct {
	Key    string          `json:"key"`
	Points []TimelinePoint `json:"points"`
}

type IngestionTimelineResponse struct {
	Status   string           `json:"status"`
	GroupBy  string           `json:"groupBy,omitempty"`
	Interval string           `json:"interval"`
	From     string           `json:"from"`
	To       string           `json:"to"`
	Points   []TimelinePoint  `json:"points,omitempty"`
	Series   []TimelineSeries `json:"series,omitempty"`
}
