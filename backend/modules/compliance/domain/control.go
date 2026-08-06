package domain

const (
	ScopeData       = "data"       // technical — evaluated via checks and/or rule coverage
	ScopeGovernance = "governance" // policy/process — not provable from logs; out of scope, excluded from score
)

const (
	StrategyAll = "ALL"
	StrategyAny = "ANY"
)

const (
	RuleNoHitsAllowed   = "NO_HITS_ALLOWED"   // pass if the query returns 0
	RuleMinHitsRequired = "MIN_HITS_REQUIRED" // pass if count >= ruleValue
	RuleThresholdMax    = "THRESHOLD_MAX"     // pass if count <= ruleValue
	RuleMatchFieldValue = "MATCH_FIELD_VALUE" // pass if `field` equals `expected`
)

type Check struct {
	Key  string `yaml:"key" json:"key"`
	Name string `yaml:"name" json:"name"`
	// DataType names the CH dataType this check reads (e.g. "wineventlog",
	// "o365"). Empty means the check spans every type in its dataset. Replaces
	// the OpenSearch indexPattern the v11 checks carried — CH has no notion of
	// an index pattern, only (dataset, dataType) scopes.
	DataType  string `yaml:"dataType,omitempty" json:"dataType,omitempty"`
	SQL       string `yaml:"sql,omitempty" json:"sql,omitempty"`
	Rule      string `yaml:"rule,omitempty" json:"rule,omitempty"`
	RuleValue *int   `yaml:"ruleValue,omitempty" json:"ruleValue,omitempty"`
	Field     string `yaml:"field,omitempty" json:"field,omitempty"`       // MATCH_FIELD_VALUE
	Expected  string `yaml:"expected,omitempty" json:"expected,omitempty"` // MATCH_FIELD_VALUE
	Todo      bool   `yaml:"todo,omitempty" json:"todo,omitempty"`         // placeholder — not yet evaluated (Pending)
}

type Control struct {
	ID          string  `yaml:"id" json:"id"`
	Family      string  `yaml:"family,omitempty" json:"family,omitempty"`
	FamilyName  string  `yaml:"familyName,omitempty" json:"familyName,omitempty"`
	Name        string  `yaml:"name" json:"name"`
	Scope       string  `yaml:"scope,omitempty" json:"scope"` // data | governance (default data)
	Statement   string  `yaml:"statement,omitempty" json:"statement,omitempty"`
	Remediation string  `yaml:"remediation,omitempty" json:"remediation,omitempty"`
	Strategy    string  `yaml:"strategy,omitempty" json:"strategy,omitempty"` // ALL | ANY (default ALL)
	Checks      []Check `yaml:"checks,omitempty" json:"checks,omitempty"`
	Source      string  `yaml:"source,omitempty" json:"source,omitempty"`

	RelPath string `yaml:"-" json:"relPath,omitempty"`
	System  bool   `yaml:"-" json:"system"`
	Enabled bool   `yaml:"-" json:"enabled"`
	Locked  bool   `yaml:"-" json:"locked"`
}

func (c *Control) EffectiveScope() string {
	if c.Scope == ScopeGovernance {
		return ScopeGovernance
	}
	return ScopeData
}

func (c *Control) EffectiveStrategy() string {
	if c.Strategy == StrategyAny {
		return StrategyAny
	}
	return StrategyAll
}

func (c *Control) RunnableChecks() []Check {
	out := make([]Check, 0, len(c.Checks))
	for _, ch := range c.Checks {
		if !ch.Todo && ch.SQL != "" {
			out = append(out, ch)
		}
	}
	return out
}
