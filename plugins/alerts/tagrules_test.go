package main

import "testing"

// ALR-F1: every operator the tagging-rule UI offers must be understood with
// the correct (not inverted) semantics — in particular, DOES_NOT_CONTAIN must
// never match a value that does contain the needle, which is exactly what the
// old default-to-"contains" fallback did.
func TestFilterTypeMatchesOperators(t *testing.T) {
	doc := map[string]any{
		"adversary": map[string]any{
			"ip":   "203.0.113.7",
			"host": "attacker-host",
			"asn":  "64500",
		},
	}

	cases := []struct {
		name  string
		f     FilterType
		field string
		want  bool
	}{
		{"IS matches", FilterType{Field: "adversary.ip", Operator: opEquals, Value: "203.0.113.7"}, "", true},
		{"IS_NOT true when different", FilterType{Field: "adversary.ip", Operator: opNotEquals, Value: "1.2.3.4"}, "", true},
		{"CONTAIN matches substring", FilterType{Field: "adversary.host", Operator: opContain, Value: "attacker"}, "", true},
		{"DOES_NOT_CONTAIN false when value contains needle (the ALR-F1 bug)", FilterType{Field: "adversary.host", Operator: opDoesNotContain, Value: "attacker"}, "", false},
		{"DOES_NOT_CONTAIN true when value lacks needle", FilterType{Field: "adversary.host", Operator: opDoesNotContain, Value: "victim"}, "", true},
		{"START_WITH matches prefix", FilterType{Field: "adversary.host", Operator: opStartWith, Value: "attacker"}, "", true},
		{"NOT_START_WITH false on matching prefix", FilterType{Field: "adversary.host", Operator: opNotStartWith, Value: "attacker"}, "", false},
		{"ENDS_WITH matches suffix", FilterType{Field: "adversary.host", Operator: opEndsWith, Value: "host"}, "", true},
		{"NOT_ENDS_WITH false on matching suffix", FilterType{Field: "adversary.host", Operator: opNotEndsWith, Value: "host"}, "", false},
		{"IS_ONE_OF matches a candidate", FilterType{Field: "adversary.ip", Operator: opIsOneOf, Value: []any{"1.2.3.4", "203.0.113.7"}}, "", true},
		{"IS_NOT_ONE_OF false when value is in the list", FilterType{Field: "adversary.ip", Operator: opIsNotOneOf, Value: []any{"203.0.113.7"}}, "", false},
		{"CONTAIN_ONE_OF matches any candidate substring", FilterType{Field: "adversary.host", Operator: opContainOneOf, Value: []any{"nope", "attack"}}, "", true},
		{"DOES_NOT_CONTAIN_ONE_OF false when one candidate matches", FilterType{Field: "adversary.host", Operator: opDoesNotContainOneOf, Value: []any{"attack"}}, "", false},
		{"IS_BETWEEN numeric range matches", FilterType{Field: "adversary.asn", Operator: opIsBetween, Value: []any{"1", "100000"}}, "", true},
		{"IS_BETWEEN numeric range excludes", FilterType{Field: "adversary.asn", Operator: opIsBetween, Value: []any{"1", "100"}}, "", false},
		{"IS_NOT_BETWEEN false when inside range", FilterType{Field: "adversary.asn", Operator: opIsNotBetween, Value: []any{"1", "100000"}}, "", false},
		{"EXIST true for present field", FilterType{Field: "adversary.ip", Operator: opExists}, "", true},
		{"DOES_NOT_EXIST true for absent field", FilterType{Field: "adversary.missing", Operator: opNotExists}, "", true},
		{"unknown operator fails closed, never matches", FilterType{Field: "adversary.host", Operator: "SOMETHING_NEW"}, "", false},
		// Legacy aliases a rule saved under the old vocabulary may still use.
		{"legacy NOT_CONTAINS behaves like DOES_NOT_CONTAIN", FilterType{Field: "adversary.host", Operator: opNotContainsLegacy, Value: "attacker"}, "", false},
		{"legacy IS_ONE_OF_TERMS behaves like IS_ONE_OF", FilterType{Field: "adversary.ip", Operator: opInLegacy, Value: []any{"203.0.113.7"}}, "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.f.matches(doc); got != tc.want {
				t.Fatalf("matches() = %v, want %v", got, tc.want)
			}
		})
	}
}
