package dto

import (
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
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
	// ID must be nil on create.
	ID *int64 `json:"id"`

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
	// ID is required on update.
	ID *int64 `json:"id"`

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

type CorrelationRuleResponse struct {
	ID int64 `json:"id"`

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

	RuleLastUpdate *time.Time `json:"ruleLastUpdate"`
	RuleActive     bool       `json:"ruleActive"`
	SystemOwner    bool       `json:"systemOwner"`

	DataTypes []RuleDataTypeResponse `json:"dataTypes"`
}

type CorrelationRuleFilters struct {
	// Page is 0-based (matches Java Spring Pageable).
	Page int `form:"page"`
	Size int `form:"size"`

	// ruleName.contains — case-insensitive partial match.
	RuleName string `form:"ruleName"`

	// ruleActive.in — optional boolean filter.
	RuleActive *bool `form:"ruleActive"`

	// ruleCategory.in — exact match values.
	RuleCategory []string `form:"ruleCategory"`

	// ruleAdversary.in — origin|target.
	RuleAdversary []string `form:"ruleAdversary"`

	// ruleTechnique.in — exact match values.
	RuleTechnique []string `form:"ruleTechnique"`

	// ruleConfidentiality.in — 0-3 values.
	RuleConfidentiality []int `form:"ruleConfidentiality"`

	// ruleIntegrity.in — 0-3 values.
	RuleIntegrity []int `form:"ruleIntegrity"`

	// ruleAvailability.in — 0-3 values.
	RuleAvailability []int `form:"ruleAvailability"`

	// systemOwner.in — optional boolean filter.
	SystemOwner *bool `form:"systemOwner"`

	// dataTypes.in — filter by associated data type strings.
	DataTypes []string `form:"dataTypes"`

	// ruleLastUpdate date range (ISO-8601 strings; matched inclusive).
	InitDate string `form:"initDate"`
	EndDate  string `form:"endDate"`

	// search — general text search against rule_name.
	Search string `form:"search"`
}

type ActivateDeactivateRequest struct {
	ID     int64 `json:"id"`
	Active bool  `json:"active"`
}

// ── Mappers ───────────────────────────────────────────────────────────────────

func rawOrEmpty(s string) json.RawMessage {
	if s == "" {
		return nil
	}
	return json.RawMessage(s)
}

func rawToString(r json.RawMessage) string {
	if len(r) == 0 {
		return ""
	}
	return string(r)
}

func dataTypeToResponse(dt domain.UtmDataTypes) RuleDataTypeResponse {
	return RuleDataTypeResponse{
		ID:                  dt.ID,
		DataType:            dt.DataType,
		DataTypeName:        dt.DataTypeName,
		DataTypeDescription: dt.DataTypeDescription,
		LastUpdate:          dt.LastUpdate,
		Included:            dt.Included,
		SystemOwner:         dt.SystemOwner,
	}
}

func CorrelationRuleToResponse(r *domain.UtmCorrelationRules) *CorrelationRuleResponse {
	resp := &CorrelationRuleResponse{
		ID:                  r.ID,
		RuleName:            r.RuleName,
		RuleAdversary:       r.RuleAdversary,
		RuleConfidentiality: r.RuleConfidentiality,
		RuleIntegrity:       r.RuleIntegrity,
		RuleAvailability:    r.RuleAvailability,
		RuleCategory:        r.RuleCategory,
		RuleTechnique:       r.RuleTechnique,
		RuleDescription:     r.RuleDescription,
		RuleReferencesDef:   rawOrEmpty(r.RuleReferencesDef),
		RuleDefinitionDef:   rawOrEmpty(r.RuleDefinitionDef),
		AfterEventsDef:      rawOrEmpty(r.AfterEventsDef),
		RuleGroupByDef:      rawOrEmpty(r.RuleGroupByDef),
		DeduplicateByDef:    rawOrEmpty(r.DeduplicateByDef),
		RuleLastUpdate:      r.RuleLastUpdate,
		RuleActive:          r.RuleActive,
		SystemOwner:         r.SystemOwner,
	}
	for _, dt := range r.DataTypes {
		resp.DataTypes = append(resp.DataTypes, dataTypeToResponse(dt))
	}
	return resp
}
