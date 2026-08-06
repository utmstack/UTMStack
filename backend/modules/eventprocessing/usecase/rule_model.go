package usecase

import (
	"errors"
	"time"
)

// Sentinel errors surfaced by the RuleStore. The usecase layer maps these onto
// the domain errors returned to callers (see mapStoreErr).
var (
	// ErrRuleNotFound is returned when a relPath does not resolve to a rule.
	ErrRuleNotFound = errors.New("rule not found")

	// ErrSystemRuleContent is returned when a write/delete is attempted against
	// a system-owned rule (system rules are read-only; only their enabled state
	// can be toggled).
	ErrSystemRuleContent = errors.New("system rule content is read-only")
)

// Impact mirrors the on-disk rule `impact` block (each score 0-3).
type Impact struct {
	Confidentiality int `yaml:"confidentiality"`
	Integrity       int `yaml:"integrity"`
	Availability    int `yaml:"availability"`
}

// Rule is the on-disk YAML representation of a correlation rule — the exact
// shape the Event Processor's cel plugin consumes. Identity is the file path
// (relPath), so there is deliberately no id field. Field order matches the
// canonical system rule files for readable diffs.
type Rule struct {
	DataTypes     []string `yaml:"dataTypes"`
	Name          string   `yaml:"name"`
	Impact        Impact   `yaml:"impact"`
	Category      string   `yaml:"category"`
	Technique     string   `yaml:"technique"`
	Adversary     string   `yaml:"adversary"`
	References    []any    `yaml:"references,omitempty"`
	Description   string   `yaml:"description,omitempty"`
	Where         string   `yaml:"where"`
	Correlation   any      `yaml:"correlation,omitempty"`
	GroupBy       []string `yaml:"groupBy,omitempty"`
	DeduplicateBy []string `yaml:"deduplicateBy,omitempty"`
	TenantId      string   `yaml:"tenantId,omitempty"`
}

// StoredRule is a Rule plus the metadata the store derives from its backing
// file and tenants.yaml: the relPath identity, whether it lives in the system
// overlay (read-only) and whether it is enabled (per tenants.yaml's
// disabledRules, keyed by the rule's basename identity — see ruleIdentity),
// and the file modification time.
type StoredRule struct {
	Rule

	RelPath  string
	Modified time.Time

	system  bool
	enabled bool
}

// Active reports whether the rule is enabled (absent from tenants.yaml's disabledRules).
func (sr *StoredRule) Active() bool { return sr.enabled }

// SystemOwned reports whether the rule lives in the system overlay.
func (sr *StoredRule) SystemOwned() bool { return sr.system }
