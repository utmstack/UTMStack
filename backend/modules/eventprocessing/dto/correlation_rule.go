package dto

import (
	"encoding/json"
	"time"
)

type DataTypeRef struct {
	ID                  *int64 `json:"id"`
	DataType            string `json:"dataType"`
	DataTypeName        string `json:"dataTypeName"`
	DataTypeDescription string `json:"dataTypeDescription"`
	Included            bool   `json:"included"`
}

type RuleDataTypeResponse struct {
	ID                  int64      `json:"id"`
	DataType            string     `json:"dataType"`
	DataTypeName        string     `json:"dataTypeName"`
	DataTypeDescription string     `json:"dataTypeDescription"`
	LastUpdate          *time.Time `json:"lastUpdate"`
	Included            bool       `json:"included"`
	SystemOwner         bool       `json:"systemOwner"`
}

type CreateCorrelationRuleRequest struct {
	// JSON tags match Java UtmCorrelationRulesDTO field names for wire compatibility.
	RuleName      string `json:"name"`
	RuleAdversary string `json:"adversary"`

	RuleConfidentiality int `json:"confidentiality"`
	RuleIntegrity       int `json:"integrity"`
	RuleAvailability    int `json:"availability"`

	RuleCategory    string `json:"category"`
	RuleTechnique   string `json:"technique"`
	RuleDescription string `json:"description"`

	RuleReferencesDef json.RawMessage `json:"references"`
	RuleDefinitionDef json.RawMessage `json:"definition"`
	AfterEventsDef    json.RawMessage `json:"afterEvents"`
	RuleGroupByDef    json.RawMessage `json:"groupBy"`
	DeduplicateByDef  json.RawMessage `json:"deduplicateBy"`

	RuleActive bool `json:"ruleActive"`

	DataTypes []DataTypeRef `json:"dataTypes"`
}

type UpdateCorrelationRuleRequest struct {
	// RelPath identifies the rule to update (the YAML-direct identity).
	RelPath string `json:"relPath"`

	RuleName      string `json:"name"`
	RuleAdversary string `json:"adversary"`

	RuleConfidentiality int `json:"confidentiality"`
	RuleIntegrity       int `json:"integrity"`
	RuleAvailability    int `json:"availability"`

	RuleCategory    string `json:"category"`
	RuleTechnique   string `json:"technique"`
	RuleDescription string `json:"description"`

	RuleReferencesDef json.RawMessage `json:"references"`
	RuleDefinitionDef json.RawMessage `json:"definition"`
	AfterEventsDef    json.RawMessage `json:"afterEvents"`
	RuleGroupByDef    json.RawMessage `json:"groupBy"`
	DeduplicateByDef  json.RawMessage `json:"deduplicateBy"`

	RuleActive bool `json:"ruleActive"`

	DataTypes []DataTypeRef `json:"dataTypes"`
}

type CorrelationRuleResponse struct {
	// RelPath is the rule identity (replaces the legacy numeric id).
	RelPath string `json:"relPath"`

	RuleName      string `json:"name"`
	RuleAdversary string `json:"adversary"`

	RuleConfidentiality int `json:"confidentiality"`
	RuleIntegrity       int `json:"integrity"`
	RuleAvailability    int `json:"availability"`

	RuleCategory    string `json:"category"`
	RuleTechnique   string `json:"technique"`
	RuleDescription string `json:"description"`

	RuleReferencesDef json.RawMessage `json:"references"`
	RuleDefinitionDef json.RawMessage `json:"definition"`
	AfterEventsDef    json.RawMessage `json:"afterEvents"`
	RuleGroupByDef    json.RawMessage `json:"groupBy"`
	DeduplicateByDef  json.RawMessage `json:"deduplicateBy"`

	RuleLastUpdate *time.Time `json:"ruleLastUpdate"`
	RuleActive     bool       `json:"ruleActive"`
	SystemOwner    bool       `json:"systemOwner"`

	DataTypes []RuleDataTypeResponse `json:"dataTypes"`
}

type CorrelationRuleFilters struct {
	// Page is 0-based (matches Java Spring Pageable).
	Page int `form:"page"`
	Size int `form:"size"`

	RuleName            string   `form:"ruleName"`            // case-insensitive partial
	RuleActive          *bool    `form:"ruleActive"`          // optional boolean
	RuleCategory        []string `form:"ruleCategory"`        // in
	RuleAdversary       []string `form:"ruleAdversary"`       // origin|target
	RuleTechnique       []string `form:"ruleTechnique"`       // in
	RuleConfidentiality []int    `form:"ruleConfidentiality"` // 0-3
	RuleIntegrity       []int    `form:"ruleIntegrity"`       // 0-3
	RuleAvailability    []int    `form:"ruleAvailability"`    // 0-3
	SystemOwner         *bool    `form:"systemOwner"`         // optional boolean
	DataTypes           []string `form:"dataTypes"`           // in
	InitDate            string   `form:"initDate"`            // ISO-8601, inclusive
	EndDate             string   `form:"endDate"`
	Search              string   `form:"search"` // general text against the name
}
