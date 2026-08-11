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
	RuleName            string          `json:"name"`
	RuleAdversary       string          `json:"adversary"`
	RuleConfidentiality int             `json:"confidentiality"`
	RuleIntegrity       int             `json:"integrity"`
	RuleAvailability    int             `json:"availability"`
	RuleCategory        string          `json:"category"`
	RuleTechnique       string          `json:"technique"`
	RuleDescription     string          `json:"description"`
	RuleReferencesDef   json.RawMessage `json:"references"`
	RuleDefinitionDef   json.RawMessage `json:"definition"`
	RuleGroupByDef      json.RawMessage `json:"groupBy"`
	DeduplicateByDef    json.RawMessage `json:"deduplicateBy"`
	RuleActive          bool            `json:"ruleActive"`
	CorrelationDef      json.RawMessage `json:"correlation"`
	DataTypes           []DataTypeRef   `json:"dataTypes"`
}

type UpdateCorrelationRuleRequest struct {
	RelPath             string          `json:"relPath"`
	RuleName            string          `json:"name"`
	RuleAdversary       string          `json:"adversary"`
	RuleConfidentiality int             `json:"confidentiality"`
	RuleIntegrity       int             `json:"integrity"`
	RuleAvailability    int             `json:"availability"`
	RuleCategory        string          `json:"category"`
	RuleTechnique       string          `json:"technique"`
	RuleDescription     string          `json:"description"`
	RuleReferencesDef   json.RawMessage `json:"references"`
	RuleDefinitionDef   json.RawMessage `json:"definition"`
	RuleGroupByDef      json.RawMessage `json:"groupBy"`
	DeduplicateByDef    json.RawMessage `json:"deduplicateBy"`
	RuleActive          bool            `json:"ruleActive"`
	CorrelationDef      json.RawMessage `json:"correlation"`
	DataTypes           []DataTypeRef   `json:"dataTypes"`
}

type CorrelationRuleResponse struct {
	RelPath             string                 `json:"relPath"`
	RuleName            string                 `json:"name"`
	RuleAdversary       string                 `json:"adversary"`
	RuleConfidentiality int                    `json:"confidentiality"`
	RuleIntegrity       int                    `json:"integrity"`
	RuleAvailability    int                    `json:"availability"`
	RuleCategory        string                 `json:"category"`
	RuleTechnique       string                 `json:"technique"`
	RuleDescription     string                 `json:"description"`
	RuleReferencesDef   json.RawMessage        `json:"references"`
	RuleDefinitionDef   json.RawMessage        `json:"definition"`
	CorrelationDef      json.RawMessage        `json:"correlation"`
	RuleGroupByDef      json.RawMessage        `json:"groupBy"`
	DeduplicateByDef    json.RawMessage        `json:"deduplicateBy"`
	RuleLastUpdate      *time.Time             `json:"ruleLastUpdate"`
	RuleActive          bool                   `json:"ruleActive"`
	SystemOwner         bool                   `json:"systemOwner"`
	DataTypes           []RuleDataTypeResponse `json:"dataTypes"`
}

type CorrelationRuleFilters struct {
	Page                int      `form:"page"`
	Size                int      `form:"size"`
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

type ImportRuleFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type ImportCorrelationRulesRequest struct {
	Files []ImportRuleFile `json:"files" binding:"required,min=1"`
}

type ImportRuleResult struct {
	Filename string `json:"filename"`
	Approved bool   `json:"approved"`
	Name     string `json:"name,omitempty"`    // rule name when parsed
	RelPath  string `json:"relPath,omitempty"` // created identity when approved
	Error    string `json:"error,omitempty"`   // reason when rejected
}

type ImportCorrelationRulesResponse struct {
	Results  []ImportRuleResult `json:"results"`
	Approved int                `json:"approved"`
	Rejected int                `json:"rejected"`
}

type ExportCorrelationRulesRequest struct {
	RelPaths []string `json:"relPaths"`
}

type ExportedRuleFile struct {
	Filename string
	Content  []byte
}
