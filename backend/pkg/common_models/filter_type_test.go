package common_models

import "testing"

func TestFilterType_ToSQLWhere(t *testing.T) {
	cases := []struct {
		name string
		in   FilterType
		want string
	}{
		{"is", FilterType{Field: "host", Operator: OpEquals, Value: "srv1"}, "host = 'srv1'"},
		{"is_not", FilterType{Field: "host", Operator: OpNotEquals, Value: "srv1"}, "host <> 'srv1'"},
		{"contain", FilterType{Field: "msg", Operator: OpContain, Value: "err"}, "msg LIKE '%err%'"},
		{"does_not_contain", FilterType{Field: "msg", Operator: OpDoesNotContain, Value: "err"}, "msg NOT LIKE '%err%'"},
		{"not_contains_alias", FilterType{Field: "msg", Operator: OpNotContains, Value: "err"}, "msg NOT LIKE '%err%'"},
		{"start_with", FilterType{Field: "path", Operator: OpStartWith, Value: "/var"}, "path LIKE '/var%'"},
		{"not_start_with", FilterType{Field: "path", Operator: OpNotStartWith, Value: "/var"}, "path NOT LIKE '/var%'"},
		{"ends_with", FilterType{Field: "path", Operator: OpEndsWith, Value: ".log"}, "path LIKE '%.log'"},
		{"not_ends_with", FilterType{Field: "path", Operator: OpNotEndsWith, Value: ".log"}, "path NOT LIKE '%.log'"},
		{"greater", FilterType{Field: "size", Operator: OpGreater, Value: "100"}, "size > '100'"},
		{"less_or_eq", FilterType{Field: "size", Operator: OpLessOrEq, Value: "100"}, "size <= '100'"},
		{"exist", FilterType{Field: "user", Operator: OpExists}, "user IS NOT NULL"},
		{"not_exist", FilterType{Field: "user", Operator: OpNotExists}, "user IS NULL"},
		{"between", FilterType{Field: "size", Operator: OpIsBetween, Value: []string{"1", "9"}}, "size BETWEEN '1' AND '9'"},
		{"not_between", FilterType{Field: "size", Operator: OpIsNotBetween, Value: []string{"1", "9"}}, "size NOT BETWEEN '1' AND '9'"},
		{"in", FilterType{Field: "host", Operator: OpIn, Value: []string{"a", "b"}}, "host IN ('a','b')"},
		{"in_or", FilterType{Field: "host", Operator: OpInOr, Value: []string{"a", "b"}}, "host IN ('a','b')"},
		{"is_one_of", FilterType{Field: "host", Operator: OpIsOneOf, Value: []string{"a", "b"}}, "host IN ('a','b')"},
		{"is_not_one_of", FilterType{Field: "host", Operator: OpIsNotOneOf, Value: []string{"a", "b"}}, "host NOT IN ('a','b')"},
		{"contain_one_of", FilterType{Field: "msg", Operator: OpContainOneOf, Value: []string{"err", "warn"}}, "(msg LIKE '%err%' OR msg LIKE '%warn%')"},
		{"does_not_contain_one_of", FilterType{Field: "msg", Operator: OpDoesNotContainOneOf, Value: []string{"a", "b"}}, "(msg NOT LIKE '%a%' AND msg NOT LIKE '%b%')"},
		{"is_in_fields_skipped", FilterType{Field: "any", Operator: OpIsInFields, Value: "x"}, ""},
		{"empty_field", FilterType{Field: "", Operator: OpEquals, Value: "x"}, ""},
		{"empty_value_scalar", FilterType{Field: "host", Operator: OpEquals, Value: ""}, ""},
		{"empty_slice_in", FilterType{Field: "host", Operator: OpIn, Value: []string{}}, ""},
		{"quote_escape", FilterType{Field: "msg", Operator: OpEquals, Value: "O'Brien"}, "msg = 'O''Brien'"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in.ToSQLWhere()
			if got != tc.want {
				t.Errorf("ToSQLWhere() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFiltersToSQLWhere(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := FiltersToSQLWhere(nil); got != "" {
			t.Errorf("nil = %q, want empty", got)
		}
		if got := FiltersToSQLWhere([]FilterType{}); got != "" {
			t.Errorf("empty = %q, want empty", got)
		}
	})
	t.Run("skips empties", func(t *testing.T) {
		got := FiltersToSQLWhere([]FilterType{
			{Field: "", Operator: OpEquals, Value: "x"},
			{Field: "host", Operator: OpEquals, Value: "srv1"},
			{Field: "user", Operator: OpEquals, Value: ""},
		})
		want := "host = 'srv1' AND "
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("multiple joined with AND and trailing AND", func(t *testing.T) {
		got := FiltersToSQLWhere([]FilterType{
			{Field: "host", Operator: OpEquals, Value: "srv1"},
			{Field: "level", Operator: OpIn, Value: []string{"ERROR", "WARN"}},
		})
		want := "host = 'srv1' AND level IN ('ERROR','WARN') AND "
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
