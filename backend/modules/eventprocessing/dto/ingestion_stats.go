package dto

type IngestionStatsBucket struct {
	Key      string `json:"key"`
	Count    int64  `json:"count"`
	LastSeen string `json:"lastSeen,omitempty"`
}

type IngestionStatsResponse struct {
	GroupBy string                 `json:"groupBy"`
	Status  string                 `json:"status"`
	From    string                 `json:"from"`
	To      string                 `json:"to"`
	Total   int64                  `json:"total"`
	Buckets []IngestionStatsBucket `json:"buckets"`
}

type TimelinePoint struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
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
