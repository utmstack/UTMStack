package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	falsePositiveTag = "False positive"

	statusOpen      = "Open"
	statusInReview  = "In review"
	statusCompleted = "Completed"

	tagRuleOpenObservation      = "This alert has been evaluated by the tag rules engine"
	tagRuleCompletedObservation = "Status changed to completed because alert was tagged as False positive"
)

type tagDecision struct {
	Tags          []string
	RuleIDs       []string
	FalsePositive bool
}

func evaluateTagRules(doc map[string]any, rules []RuleSnapshot) tagDecision {
	var d tagDecision
	seen := make(map[string]bool)

	for _, rule := range rules {
		if !ruleMatches(rule, doc) {
			continue
		}
		d.RuleIDs = append(d.RuleIDs, rule.ID)
		for _, t := range rule.TagNames {
			if !seen[t] {
				seen[t] = true
				d.Tags = append(d.Tags, t)
			}
		}
	}

	for _, t := range d.Tags {
		if t == falsePositiveTag {
			d.FalsePositive = true
			break
		}
	}
	return d
}

func ruleMatches(rule RuleSnapshot, doc map[string]any) bool {
	if len(rule.Conditions) == 0 {
		return false
	}
	for _, c := range rule.Conditions {
		if !c.matches(doc) {
			return false
		}
	}
	return true
}

type FilterOperator string

const (
	opEquals      FilterOperator = "IS"
	opNotEquals   FilterOperator = "IS_NOT"
	opIn          FilterOperator = "IS_ONE_OF_TERMS"
	opInOr        FilterOperator = "IS_ONE_OF_TERMS_OR"
	opExists      FilterOperator = "EXIST"
	opNotExists   FilterOperator = "DOES_NOT_EXIST"
	opNotContains FilterOperator = "NOT_CONTAINS"
)

type FilterType struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    any            `json:"value,omitempty"`
}

func (f FilterType) matches(doc map[string]any) bool {
	field := strings.TrimSuffix(f.Field, ".keyword")
	values := collectValues(doc, strings.Split(field, "."))

	switch f.Operator {
	case opExists:
		return len(values) > 0
	case opNotExists:
		return len(values) == 0
	case opEquals:
		return anyMatch(values, func(v any) bool { return valueEquals(v, f.Value) })
	case opNotEquals:
		// must_not term: true when no value equals (including an absent field).
		return !anyMatch(values, func(v any) bool { return valueEquals(v, f.Value) })
	case opIn, opInOr:
		cands := toAnySlice(f.Value)
		return anyMatch(values, func(v any) bool {
			for _, c := range cands {
				if valueEquals(v, c) {
					return true
				}
			}
			return false
		})
	case opNotContains:
		return !anyMatch(values, func(v any) bool { return valueContains(v, f.Value) })
	default: // analyzed match fallback (approximated)
		return anyMatch(values, func(v any) bool { return valueContains(v, f.Value) })
	}
}

// collectValues resolves a dotted path into every leaf value it reaches,
// branching over array elements at each step (OpenSearch-style flattening).
func collectValues(node any, segments []string) []any {
	if node == nil {
		return nil
	}
	if arr, ok := node.([]any); ok {
		var out []any
		for _, e := range arr {
			out = append(out, collectValues(e, segments)...)
		}
		return out
	}
	if len(segments) == 0 {
		return []any{node}
	}
	if obj, ok := node.(map[string]any); ok {
		child, present := obj[segments[0]]
		if !present {
			return nil
		}
		return collectValues(child, segments[1:])
	}
	// More path to walk but the node is a scalar — no match.
	return nil
}

func anyMatch(values []any, pred func(any) bool) bool {
	for _, v := range values {
		if pred(v) {
			return true
		}
	}
	return false
}

// valueEquals compares a leaf value (from encoding/json: string, float64, bool)
// to a condition value of the same shape, tolerating string/number formatting.
func valueEquals(v, cond any) bool {
	switch c := cond.(type) {
	case string:
		return toString(v) == c
	case float64:
		if n, ok := v.(float64); ok {
			return n == c
		}
		return toString(v) == strconv.FormatFloat(c, 'f', -1, 64)
	case bool:
		if b, ok := v.(bool); ok {
			return b == c
		}
		return toString(v) == strconv.FormatBool(c)
	default:
		return toString(v) == toString(cond)
	}
}

func valueContains(v, cond any) bool {
	return strings.Contains(strings.ToLower(toString(v)), strings.ToLower(toString(cond)))
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

func toAnySlice(v any) []any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []any:
		return t
	case []string:
		out := make([]any, len(t))
		for i, s := range t {
			out[i] = s
		}
		return out
	default:
		return []any{v}
	}
}
