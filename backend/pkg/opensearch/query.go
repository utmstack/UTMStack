package opensearch

// This file contains convenience helpers for building OpenSearch query maps.
// All functions return plain map[string]any so callers can compose them freely.

// TermQuery builds a { "term": { field: value } } clause.
func TermQuery(field string, value any) map[string]any {
	return map[string]any{
		"term": map[string]any{
			field: value,
		},
	}
}

// TermsQuery builds a { "terms": { field: values } } clause.
func TermsQuery(field string, values []string) map[string]any {
	iValues := make([]any, len(values))
	for i, v := range values {
		iValues[i] = v
	}
	return map[string]any{
		"terms": map[string]any{
			field: iValues,
		},
	}
}

// MustNotExistsQuery builds an exists query suitable for use in must_not.
func MustNotExistsQuery(field string) map[string]any {
	return map[string]any{
		"exists": map[string]any{
			"field": field,
		},
	}
}

// BoolQuery builds a bool compound query. Pass nil for any clause you don't need.
func BoolQuery(must, filter, should, mustNot []map[string]any) map[string]any {
	b := map[string]any{}
	if len(must) > 0 {
		b["must"] = must
	}
	if len(filter) > 0 {
		b["filter"] = filter
	}
	if len(should) > 0 {
		b["should"] = should
	}
	if len(mustNot) > 0 {
		b["must_not"] = mustNot
	}
	return map[string]any{"bool": b}
}

// MatchQuery builds a { "match": { field: value } } clause.
func MatchQuery(field string, value any) map[string]any {
	return map[string]any{
		"match": map[string]any{
			field: value,
		},
	}
}
