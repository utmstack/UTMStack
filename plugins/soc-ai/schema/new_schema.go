package schema

import "github.com/threatwinds/go-sdk/plugins"

type AlertFields struct {
	Timestamp         string           `json:"@timestamp"`
	ID                string           `json:"id"`
	ParentID          *string          `json:"parentId,omitempty"`
	Status            int              `json:"status"`
	StatusLabel       string           `json:"statusLabel"`
	StatusObservation string           `json:"statusObservation"`
	IsIncident        bool             `json:"isIncident"`
	IncidentDetail    IncidentDetail   `json:"incidentDetail"`
	Name              string           `json:"name"`
	Category          string           `json:"category"`
	Severity          int              `json:"severity"`
	SeverityLabel     string           `json:"severityLabel"`
	Description       string           `json:"description"`
	Solution          string           `json:"solution"`
	Technique         string           `json:"technique"`
	Reference         []string         `json:"reference"`
	DataType          string           `json:"dataType"`
	Impact            *plugins.Impact  `json:"impact"`
	ImpactScore       uint32           `json:"impactScore"`
	DataSource        string           `json:"dataSource"`
	Adversary         *plugins.Side    `json:"adversary"`
	Target            *plugins.Side    `json:"target"`
	Events            []*plugins.Event `json:"events,omitempty"`
	LastEvent         *plugins.Event   `json:"lastEvent"`
	Tags              []string         `json:"tags"`
	Notes             string           `json:"notes"`
	TagRulesApplied   []int            `json:"tagRulesApplied,omitempty"`
	DeduplicatedBy    []string         `json:"deduplicatedBy,omitempty"`
	GroupedBy         []string         `json:"groupedBy,omitempty"`
	GPTTimestamp      string           `json:"gpt_timestamp,omitempty"`
	GPTClassification string           `json:"gpt_classification,omitempty"`
	GPTReasoning      string           `json:"gpt_reasoning,omitempty"`
	GPTNextSteps      string           `json:"gpt_next_steps,omitempty"`
}

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	CreationDate string `json:"creationDate"`
	Source       string `json:"source"`
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
