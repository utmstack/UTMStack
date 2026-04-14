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
	Status            int            `json:"status"`
	StatusLabel       string         `json:"statusLabel"`
	StatusObservation string         `json:"statusObservation"`
	IsIncident        bool           `json:"isIncident"`
	IncidentDetail    IncidentDetail `json:"incidentDetail"`
	Severity          int            `json:"severity"`
	SeverityLabel     string         `json:"severityLabel"`
	Solution          string         `json:"solution"`
	Reference         []string       `json:"reference"`
	LastEvent         *plugins.Event `json:"lastEvent"`
	Tags              []string       `json:"tags"`
	Notes             string         `json:"notes"`
	TagRulesApplied   []int          `json:"tagRulesApplied,omitempty"`
	DeduplicatedBy    []string       `json:"deduplicatedBy,omitempty"`
	GroupedBy         []string       `json:"groupedBy,omitempty"`
	GPTTimestamp      string         `json:"gpt_timestamp,omitempty"`
	GPTClassification string         `json:"gpt_classification,omitempty"`
	GPTReasoning      string         `json:"gpt_reasoning,omitempty"`
	GPTNextSteps      string         `json:"gpt_next_steps,omitempty"`
	AnonymizedFields  []string       `json:"anonymizedFields,omitempty"` // Fields that were anonymized for privacy
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
