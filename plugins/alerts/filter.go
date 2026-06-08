package main

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

// toQueryClause renders a single FilterType as an OpenSearch clause that can
// sit inside a `must` array. Mirrors the legacy filterTypeToQuery from the
// backend's alert_os.go so query semantics stay aligned with the data the
// rules were authored against.
func (f FilterType) toQueryClause() map[string]any {
	switch f.Operator {
	case opEquals:
		return termQuery(f.Field, f.Value)
	case opNotEquals:
		return map[string]any{
			"bool": map[string]any{
				"must_not": termQuery(f.Field, f.Value),
			},
		}
	case opIn, opInOr:
		return map[string]any{
			"terms": map[string]any{
				f.Field: toAnySlice(f.Value),
			},
		}
	case opExists:
		return map[string]any{
			"exists": map[string]any{"field": f.Field},
		}
	case opNotExists:
		return map[string]any{
			"bool": map[string]any{
				"must_not": map[string]any{
					"exists": map[string]any{"field": f.Field},
				},
			},
		}
	case opNotContains:
		return map[string]any{
			"bool": map[string]any{
				"must_not": matchQuery(f.Field, f.Value),
			},
		}
	default:
		return matchQuery(f.Field, f.Value)
	}
}

func termQuery(field string, value any) map[string]any {
	return map[string]any{"term": map[string]any{field: value}}
}

func matchQuery(field string, value any) map[string]any {
	return map[string]any{"match": map[string]any{field: value}}
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
