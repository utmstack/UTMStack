package schema

import "github.com/threatwinds/go-sdk/plugins"

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	CreationDate string `json:"creationDate"`
	Source       string `json:"source"`
}

type AlertFields struct {
	Timestamp         string         `json:"@timestamp"`
	Status            string         `json:"status,omitempty"`
	StatusObservation string         `json:"statusObservation,omitempty"`
	IsIncident        bool           `json:"isIncident,omitempty"`
	IncidentDetail    IncidentDetail `json:"incidentDetail,omitzero"`
	Solution          string         `json:"solution,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	Notes             string         `json:"notes,omitempty"`
	TagRulesApplied   []string       `json:"tagRulesApplied,omitempty"`
	LastEvent         *plugins.Event `json:"lastEvent,omitempty"`
	AnonymizedFields  []string       `json:"anonymizedFields,omitempty"`
	plugins.Alert
}

type AlertCorrelation struct {
	CurrentAlert  AlertFields
	RelatedAlerts []AlertFields
	Counts        MatchTypeCounts
}

type AlertCounts struct {
	Incidents     int
	FalsePositive int
	Standard      int
	Unclassified  int
}

type MatchTypeCounts struct {
	OriginIP   AlertCounts
	TargetIP   AlertCounts
	OriginUser AlertCounts
	TargetUser AlertCounts
}
