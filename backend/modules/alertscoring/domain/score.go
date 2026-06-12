package domain

const (
	DecisionComplete = "COMPLETE"  // low risk — safe to auto-close
	DecisionInReview = "IN_REVIEW" // needs a human analyst
	DecisionIncident = "INCIDENT"  // escalate / correlate into an incident
)

type PhaseResult struct {
	Score     int            `json:"score"`
	Breakdown map[string]any `json:"breakdown,omitempty"`
}

type Score struct {
	AlertID    string                 `json:"alertId"`
	AlertName  string                 `json:"alertName"`
	FinalScore int                    `json:"finalScore"`
	Decision   string                 `json:"decision"`
	Weights    map[string]float64     `json:"weights"`
	Phases     map[string]PhaseResult `json:"phases"`
	Summary    string                 `json:"summary"`
}
