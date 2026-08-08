package domain

import "github.com/utmstack/utmstack/backend/pkg/common_models"

type ControlScope string

const (
	ScopeData       ControlScope = "data"       // technical — evaluated via checks and/or rule coverage
	ScopeGovernance ControlScope = "governance" // policy/process — not provable from logs; out of scope, excluded from score
)

type CheckStrategy string

const (
	StrategyAll CheckStrategy = "ALL"
	StrategyAny CheckStrategy = "ANY"
)

type Dataset string

const (
	DatasetLogs   Dataset = "logs"
	DatasetAlerts Dataset = "alerts"
)

type CheckRule string

const (
	RuleMinHitsRequired CheckRule = "MIN_HITS_REQUIRED" // pass if count >= ruleValue
	RuleThresholdMax    CheckRule = "THRESHOLD_MAX"     // pass if count <= ruleValue; 0 means "none allowed"
)

type Check struct {
	Key       string                     `yaml:"key" json:"key"`
	Name      string                     `yaml:"name" json:"name"`
	Dataset   Dataset                    `yaml:"dataset,omitempty" json:"dataset,omitempty"`
	DataType  string                     `yaml:"dataType,omitempty" json:"dataType,omitempty"` // wineventlog, o365, aws-cloudtrail… empty means every type in the dataset
	Filters   []common_models.FilterType `yaml:"filters,omitempty" json:"filters,omitempty"`
	Rule      CheckRule                  `yaml:"rule,omitempty" json:"rule,omitempty"`
	RuleValue *int                       `yaml:"ruleValue,omitempty" json:"ruleValue,omitempty"`
	Todo      bool                       `yaml:"todo,omitempty" json:"todo,omitempty"` // placeholder — not yet defined (Pending)
}

type Control struct {
	ID          string        `yaml:"id" json:"id"`
	Family      string        `yaml:"family,omitempty" json:"family,omitempty"`
	FamilyName  string        `yaml:"familyName,omitempty" json:"familyName,omitempty"`
	Name        string        `yaml:"name" json:"name"`
	Scope       ControlScope  `yaml:"scope,omitempty" json:"scope"`
	Statement   string        `yaml:"statement,omitempty" json:"statement,omitempty"`
	Remediation string        `yaml:"remediation,omitempty" json:"remediation,omitempty"`
	Strategy    CheckStrategy `yaml:"strategy,omitempty" json:"strategy,omitempty"`
	Checks      []Check       `yaml:"checks,omitempty" json:"checks,omitempty"`
	Source      string        `yaml:"source,omitempty" json:"source,omitempty"`
}
