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

// This vocabulary must stay exactly in sync with the frontend's RULE_OPERATORS
// (frontend/src/features/alerts/lib/tagging-rule-meta.ts) and with the
// canonical backend/pkg/common_models.FilterOperator it mirrors — this plugin
// is its own Go module and can't import that package directly, so the string
// values are duplicated here on purpose. ALR-F1: this list used to recognize
// only 4 of the 16 operators the tagging-rule UI actually offers; everything
// else silently fell through to a "contains" approximation, which inverted
// every negated operator (DOES_NOT_CONTAIN, IS_NOT_ONE_OF, NOT_START_WITH,
// NOT_ENDS_WITH, IS_NOT_BETWEEN) — "does not contain X" ended up tagging
// exactly the alerts that DO contain X.
const (
	opEquals              FilterOperator = "IS"
	opNotEquals           FilterOperator = "IS_NOT"
	opContain             FilterOperator = "CONTAIN"
	opDoesNotContain      FilterOperator = "DOES_NOT_CONTAIN"
	opStartWith           FilterOperator = "START_WITH"
	opNotStartWith        FilterOperator = "NOT_START_WITH"
	opEndsWith            FilterOperator = "ENDS_WITH"
	opNotEndsWith         FilterOperator = "NOT_ENDS_WITH"
	opIsOneOf             FilterOperator = "IS_ONE_OF"
	opIsNotOneOf          FilterOperator = "IS_NOT_ONE_OF"
	opContainOneOf        FilterOperator = "CONTAIN_ONE_OF"
	opDoesNotContainOneOf FilterOperator = "DOES_NOT_CONTAIN_ONE_OF"
	opIsBetween           FilterOperator = "IS_BETWEEN"
	opIsNotBetween        FilterOperator = "IS_NOT_BETWEEN"
	opExists              FilterOperator = "EXIST"
	opNotExists           FilterOperator = "DOES_NOT_EXIST"

	// Older strings a rule saved before the UI settled on the list above may
	// still carry. Kept as aliases of their current equivalent so an existing
	// rule doesn't silently stop matching after this fix.
	opInLegacy          FilterOperator = "IS_ONE_OF_TERMS"
	opInOrLegacy        FilterOperator = "IS_ONE_OF_TERMS_OR"
	opNotContainsLegacy FilterOperator = "NOT_CONTAINS"
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
	case opContain:
		return anyMatch(values, func(v any) bool { return valueContains(v, f.Value) })
	case opDoesNotContain, opNotContainsLegacy:
		return !anyMatch(values, func(v any) bool { return valueContains(v, f.Value) })
	case opStartWith:
		return anyMatch(values, func(v any) bool { return valueHasPrefix(v, f.Value) })
	case opNotStartWith:
		return !anyMatch(values, func(v any) bool { return valueHasPrefix(v, f.Value) })
	case opEndsWith:
		return anyMatch(values, func(v any) bool { return valueHasSuffix(v, f.Value) })
	case opNotEndsWith:
		return !anyMatch(values, func(v any) bool { return valueHasSuffix(v, f.Value) })
	case opIsOneOf, opInLegacy, opInOrLegacy:
		cands := toAnySlice(f.Value)
		return anyMatch(values, func(v any) bool { return valueInList(v, cands) })
	case opIsNotOneOf:
		cands := toAnySlice(f.Value)
		return !anyMatch(values, func(v any) bool { return valueInList(v, cands) })
	case opContainOneOf:
		cands := toAnySlice(f.Value)
		return anyMatch(values, func(v any) bool { return valueContainsAny(v, cands) })
	case opDoesNotContainOneOf:
		cands := toAnySlice(f.Value)
		return !anyMatch(values, func(v any) bool { return valueContainsAny(v, cands) })
	case opIsBetween:
		lo, hi, ok := betweenBounds(f.Value)
		return ok && anyMatch(values, func(v any) bool { return valueInRange(v, lo, hi) })
	case opIsNotBetween:
		lo, hi, ok := betweenBounds(f.Value)
		if !ok {
			return false
		}
		return !anyMatch(values, func(v any) bool { return valueInRange(v, lo, hi) })
	default:
		// An operator this plugin doesn't recognize must never silently match
		// (ALR-F1's bug was exactly that default-to-"contains" fallback).
		return false
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

func valueHasPrefix(v, cond any) bool {
	return strings.HasPrefix(strings.ToLower(toString(v)), strings.ToLower(toString(cond)))
}

func valueHasSuffix(v, cond any) bool {
	return strings.HasSuffix(strings.ToLower(toString(v)), strings.ToLower(toString(cond)))
}

func valueInList(v any, cands []any) bool {
	for _, c := range cands {
		if valueEquals(v, c) {
			return true
		}
	}
	return false
}

func valueContainsAny(v any, cands []any) bool {
	for _, c := range cands {
		if valueContains(v, c) {
			return true
		}
	}
	return false
}

// betweenBounds reads the two-element [min, max] list IS_BETWEEN/IS_NOT_BETWEEN
// carry (frontend/src/features/alerts/lib/tagging-rule-form.ts serializes
// "list"-need operators as a comma-split array). Unparsable bounds mean the
// condition can't be evaluated, so ok=false rather than guessing.
func betweenBounds(value any) (lo, hi float64, ok bool) {
	cands := toAnySlice(value)
	if len(cands) < 2 {
		return 0, 0, false
	}
	lo, err1 := strconv.ParseFloat(toString(cands[0]), 64)
	hi, err2 := strconv.ParseFloat(toString(cands[1]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	return lo, hi, true
}

func valueInRange(v any, lo, hi float64) bool {
	n, err := strconv.ParseFloat(toString(v), 64)
	if err != nil {
		return false
	}
	return n >= lo && n <= hi
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
