package common_models

import "strings"

// IndexTimestampFormat is the timestamp format used by OpenSearch range queries.
const IndexTimestampFormat = "yyyy-MM-dd'T'HH:mm:ss.SSSXXX"

// FilterOperator matches Java's FilterOperator enum used by tag-rule conditions
// and dashboard queries.
type FilterOperator string

const (
	OpEquals      FilterOperator = "IS"
	OpNotEquals   FilterOperator = "IS_NOT"
	OpIn          FilterOperator = "IS_ONE_OF_TERMS"
	OpInOr        FilterOperator = "IS_ONE_OF_TERMS_OR"
	OpLessOrEq    FilterOperator = "IS_LESS_THAN_OR_EQUALS"
	OpGreater     FilterOperator = "IS_GREATER_THAN"
	OpExists      FilterOperator = "EXIST"
	OpNotExists   FilterOperator = "DOES_NOT_EXIST"
	OpNotContains FilterOperator = "NOT_CONTAINS"

	// Extended operators ported from Java OperatorType enum.
	OpContain                FilterOperator = "CONTAIN"
	OpDoesNotContain         FilterOperator = "DOES_NOT_CONTAIN"
	OpContainOneOf           FilterOperator = "CONTAIN_ONE_OF"
	OpDoesNotContainOneOf    FilterOperator = "DOES_NOT_CONTAIN_ONE_OF"
	OpIsOneOf                FilterOperator = "IS_ONE_OF"
	OpIsNotOneOf             FilterOperator = "IS_NOT_ONE_OF"
	OpIsBetween              FilterOperator = "IS_BETWEEN"
	OpIsNotBetween           FilterOperator = "IS_NOT_BETWEEN"
	OpIsInFields             FilterOperator = "IS_IN_FIELDS"
	OpIsNotInFields          FilterOperator = "IS_NOT_IN_FIELDS"
	OpEndsWith               FilterOperator = "ENDS_WITH"
	OpNotEndsWith            FilterOperator = "NOT_ENDS_WITH"
	OpStartWith              FilterOperator = "START_WITH"
	OpNotStartWith           FilterOperator = "NOT_START_WITH"
)

// FilterType mirrors Java's FilterType used in alert tag-rule conditions
// and cross-module dashboard filter DSL.
type FilterType struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    any            `json:"value,omitempty"`
}

// ToQuery converts a single FilterType into an OpenSearch bool-clause fragment.
// The returned map contains a single key — "filter", "must_not", or "should" —
// whose value is the inner query map. FiltersToQuery assembles these into the
// root bool query.
func (f FilterType) ToQuery() map[string]any {
	switch f.Operator {
	// -----------------------------------------------------------------------
	// IS / IS_NOT  →  match_phrase
	// -----------------------------------------------------------------------
	case OpEquals:
		return map[string]any{
			"filter": map[string]any{
				"match_phrase": map[string]any{f.Field: f.Value},
			},
		}
	case OpNotEquals:
		return map[string]any{
			"must_not": map[string]any{
				"match_phrase": map[string]any{f.Field: f.Value},
			},
		}

	// -----------------------------------------------------------------------
	// CONTAIN / DOES_NOT_CONTAIN  →  query_string *v*
	// -----------------------------------------------------------------------
	case OpContain:
		return map[string]any{
			"filter": map[string]any{
				"query_string": map[string]any{
					"fields":           []string{f.Field},
					"query":            "*" + escapeQueryString(toString(f.Value)) + "*",
					"default_operator": "AND",
					"lenient":          true,
					"type":             "best_fields",
				},
			},
		}
	case OpDoesNotContain:
		return map[string]any{
			"must_not": map[string]any{
				"query_string": map[string]any{
					"fields":           []string{f.Field},
					"query":            "*" + escapeQueryString(toString(f.Value)) + "*",
					"default_operator": "AND",
					"lenient":          true,
					"type":             "best_fields",
				},
			},
		}

	// -----------------------------------------------------------------------
	// CONTAIN_ONE_OF  →  filter: bool{should:[match_phrase]}
	// -----------------------------------------------------------------------
	case OpContainOneOf:
		should := buildMatchPhraseShould(f.Field, toStringSlice(f.Value))
		return map[string]any{
			"filter": map[string]any{
				"bool": map[string]any{
					"should":               should,
					"minimum_should_match": 1,
				},
			},
		}

	// -----------------------------------------------------------------------
	// DOES_NOT_CONTAIN_ONE_OF  →  filter: bool{must_not:[match_phrase]}
	// (NOT top-level must_not — Java parity)
	// -----------------------------------------------------------------------
	case OpDoesNotContainOneOf:
		mustNots := buildMatchPhraseSlice(f.Field, toStringSlice(f.Value))
		return map[string]any{
			"filter": map[string]any{
				"bool": map[string]any{
					"must_not": mustNots,
				},
			},
		}

	// -----------------------------------------------------------------------
	// IS_ONE_OF  →  filter: bool{should:[match_phrase]}
	// -----------------------------------------------------------------------
	case OpIsOneOf:
		should := buildMatchPhraseShould(f.Field, toStringSlice(f.Value))
		return map[string]any{
			"filter": map[string]any{
				"bool": map[string]any{
					"should":               should,
					"minimum_should_match": 1,
				},
			},
		}

	// -----------------------------------------------------------------------
	// IS_NOT_ONE_OF  →  must_not: bool{should:[match_phrase]}
	// -----------------------------------------------------------------------
	case OpIsNotOneOf:
		should := buildMatchPhraseShould(f.Field, toStringSlice(f.Value))
		return map[string]any{
			"must_not": map[string]any{
				"bool": map[string]any{
					"should":               should,
					"minimum_should_match": 1,
				},
			},
		}

	// -----------------------------------------------------------------------
	// EXIST / DOES_NOT_EXIST  →  exists
	// -----------------------------------------------------------------------
	case OpExists:
		return map[string]any{
			"filter": map[string]any{
				"exists": map[string]any{"field": f.Field},
			},
		}
	case OpNotExists:
		return map[string]any{
			"must_not": map[string]any{
				"exists": map[string]any{"field": f.Field},
			},
		}

	// -----------------------------------------------------------------------
	// IS_BETWEEN / IS_NOT_BETWEEN  →  range gte+lte
	// -----------------------------------------------------------------------
	case OpIsBetween:
		vals := toStringSlice(f.Value)
		return map[string]any{
			"filter": map[string]any{
				"range": map[string]any{
					f.Field: map[string]any{
						"gte": safeIdx(vals, 0),
						"lte": safeIdx(vals, 1),
					},
				},
			},
		}
	case OpIsNotBetween:
		vals := toStringSlice(f.Value)
		return map[string]any{
			"must_not": map[string]any{
				"range": map[string]any{
					f.Field: map[string]any{
						"gte": safeIdx(vals, 0),
						"lte": safeIdx(vals, 1),
					},
				},
			},
		}

	// -----------------------------------------------------------------------
	// IS_IN_FIELDS  →  filter: query_string{default_field:"*", query:"*v*"}
	// IS_NOT_IN_FIELDS  →  filter: bool{must_not:query_string}
	// -----------------------------------------------------------------------
	case OpIsInFields:
		return map[string]any{
			"filter": map[string]any{
				"query_string": map[string]any{
					"default_field": "*",
					"query":         "*" + escapeQueryString(toString(f.Value)) + "*",
				},
			},
		}
	case OpIsNotInFields:
		return map[string]any{
			"filter": map[string]any{
				"bool": map[string]any{
					"must_not": map[string]any{
						"query_string": map[string]any{
							"default_field": "*",
							"query":         "*" + escapeQueryString(toString(f.Value)) + "*",
						},
					},
				},
			},
		}

	// -----------------------------------------------------------------------
	// ENDS_WITH / NOT_ENDS_WITH  →  wildcard *v
	// -----------------------------------------------------------------------
	case OpEndsWith:
		return map[string]any{
			"filter": map[string]any{
				"wildcard": map[string]any{f.Field: "*" + toString(f.Value)},
			},
		}
	case OpNotEndsWith:
		return map[string]any{
			"filter": map[string]any{
				"bool": map[string]any{
					"must_not": map[string]any{
						"wildcard": map[string]any{f.Field: "*" + toString(f.Value)},
					},
				},
			},
		}

	// -----------------------------------------------------------------------
	// START_WITH / NOT_START_WITH  →  wildcard v*
	// -----------------------------------------------------------------------
	case OpStartWith:
		return map[string]any{
			"filter": map[string]any{
				"wildcard": map[string]any{f.Field: toString(f.Value) + "*"},
			},
		}
	case OpNotStartWith:
		return map[string]any{
			"filter": map[string]any{
				"bool": map[string]any{
					"must_not": map[string]any{
						"wildcard": map[string]any{f.Field: toString(f.Value) + "*"},
					},
				},
			},
		}

	// -----------------------------------------------------------------------
	// IS_ONE_OF_TERMS  →  filter: terms
	// -----------------------------------------------------------------------
	case OpIn:
		return map[string]any{
			"filter": map[string]any{
				"terms": map[string]any{f.Field: toAnySlice(f.Value)},
			},
		}

	// -----------------------------------------------------------------------
	// IS_ONE_OF_TERMS_OR  →  should: terms  (caller sets minimum_should_match)
	// -----------------------------------------------------------------------
	case OpInOr:
		return map[string]any{
			"should": map[string]any{
				"terms": map[string]any{f.Field: toAnySlice(f.Value)},
			},
		}

	// -----------------------------------------------------------------------
	// IS_GREATER_THAN  →  filter: range{gt + format}
	// -----------------------------------------------------------------------
	case OpGreater:
		return map[string]any{
			"filter": map[string]any{
				"range": map[string]any{
					f.Field: map[string]any{
						"gt":     f.Value,
						"format": IndexTimestampFormat,
					},
				},
			},
		}

	// -----------------------------------------------------------------------
	// IS_LESS_THAN_OR_EQUALS  →  filter: range{lte + format}
	// -----------------------------------------------------------------------
	case OpLessOrEq:
		return map[string]any{
			"filter": map[string]any{
				"range": map[string]any{
					f.Field: map[string]any{
						"lte":    f.Value,
						"format": IndexTimestampFormat,
					},
				},
			},
		}

	default:
		// Fallback: treat as IS (match_phrase filter)
		return map[string]any{
			"filter": map[string]any{
				"match_phrase": map[string]any{f.Field: f.Value},
			},
		}
	}
}

// FiltersToQuery assembles a root OpenSearch bool query from a slice of FilterType.
// Each filter's ToQuery() result is routed into the correct top-level bucket
// (filter, must_not, or should). If any filter uses IS_ONE_OF_TERMS_OR,
// minimum_should_match:1 is set on the outer bool.
//
// Returns {"bool": {}} for nil/empty input.
func FiltersToQuery(filters []FilterType) map[string]any {
	if len(filters) == 0 {
		return map[string]any{"bool": map[string]any{}}
	}

	var filterClauses []any
	var mustNotClauses []any
	var shouldClauses []any
	hasOr := false

	for _, f := range filters {
		if f.Operator == OpInOr {
			hasOr = true
		}
		clause := f.ToQuery()
		if inner, ok := clause["filter"]; ok {
			filterClauses = append(filterClauses, inner)
		} else if inner, ok := clause["must_not"]; ok {
			mustNotClauses = append(mustNotClauses, inner)
		} else if inner, ok := clause["should"]; ok {
			shouldClauses = append(shouldClauses, inner)
		}
	}

	boolClause := map[string]any{}
	if len(filterClauses) > 0 {
		boolClause["filter"] = filterClauses
	}
	if len(mustNotClauses) > 0 {
		boolClause["must_not"] = mustNotClauses
	}
	if len(shouldClauses) > 0 {
		boolClause["should"] = shouldClauses
	}
	if hasOr {
		boolClause["minimum_should_match"] = 1
	}

	return map[string]any{"bool": boolClause}
}

// ---------------------------------------------------------------------------
// private helpers
// ---------------------------------------------------------------------------

// escapeQueryString escapes Lucene/OpenSearch query_string special characters.
func escapeQueryString(s string) string {
	special := `+-=&&||><!(){}[]^"~*?:\/`
	var b strings.Builder
	b.Grow(len(s) * 2)
	for _, ch := range s {
		if strings.ContainsRune(special, ch) {
			b.WriteRune('\\')
		}
		b.WriteRune(ch)
	}
	return b.String()
}

// toStringSlice coerces an any value to []string.
func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		return []string{t}
	default:
		return nil
	}
}

// toAnySlice coerces an any value to []any.
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

// toString converts an any to string with a safe fallback.
func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// safeIdx returns the element at index i from a slice, or "" if out of bounds.
func safeIdx(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return ""
}

// buildMatchPhraseShould returns a slice of match_phrase clauses for use in
// a should array (each as map[string]any).
func buildMatchPhraseShould(field string, values []string) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, v := range values {
		out = append(out, map[string]any{
			"match_phrase": map[string]any{field: v},
		})
	}
	return out
}

// buildMatchPhraseSlice returns a slice of match_phrase clauses (for must_not).
func buildMatchPhraseSlice(field string, values []string) []map[string]any {
	return buildMatchPhraseShould(field, values)
}
