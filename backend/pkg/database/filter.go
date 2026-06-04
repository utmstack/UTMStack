package database

import (
	"fmt"
	"strings"
)

// Operator is a filter comparison operator in the filter DSL / query params.
type Operator string

const (
	OpEqual       Operator = "eq"    // field = value
	OpNotEqual    Operator = "neq"   // field <> value
	OpContains    Operator = "like"  // field ILIKE %value%
	OpNotContains Operator = "nlike" // field NOT ILIKE %value%
	OpIn          Operator = "in"    // field IN (csv)
	OpNotIn       Operator = "nin"   // field NOT IN (csv)
	OpGreater     Operator = "gt"    // field > value
	OpGreaterEq   Operator = "gte"   // field >= value
	OpLess        Operator = "lt"    // field < value
	OpLessEq      Operator = "lte"   // field <= value
	OpIsNull      Operator = "null"  // field IS NULL
	OpIsNotNull   Operator = "nnull" // field IS NOT NULL
)

// Filter is a single parsed condition. Field is always allowlist-validated
// before a Filter is built from client input.
type Filter struct {
	Field    string
	Operator Operator
	Value    any
}

// ToWhereClause turns a Filter into a parameterized WHERE clause. The field name
// goes into the SQL text (column names can't be parameterized) — callers must
// only ever build Filters from allowlisted fields.
func (f Filter) ToWhereClause() WhereClause {
	switch f.Operator {
	case OpNotEqual:
		return WhereClause{Query: f.Field + " <> ?", Args: []any{f.Value}}
	case OpContains:
		return WhereClause{Query: f.Field + " ILIKE ?", Args: []any{"%" + fmt.Sprint(f.Value) + "%"}}
	case OpNotContains:
		return WhereClause{Query: f.Field + " NOT ILIKE ?", Args: []any{"%" + fmt.Sprint(f.Value) + "%"}}
	case OpIn:
		return WhereClause{Query: f.Field + " IN ?", Args: []any{f.Value}}
	case OpNotIn:
		return WhereClause{Query: f.Field + " NOT IN ?", Args: []any{f.Value}}
	case OpGreater:
		return WhereClause{Query: f.Field + " > ?", Args: []any{f.Value}}
	case OpGreaterEq:
		return WhereClause{Query: f.Field + " >= ?", Args: []any{f.Value}}
	case OpLess:
		return WhereClause{Query: f.Field + " < ?", Args: []any{f.Value}}
	case OpLessEq:
		return WhereClause{Query: f.Field + " <= ?", Args: []any{f.Value}}
	case OpIsNull:
		return WhereClause{Query: f.Field + " IS NULL"}
	case OpIsNotNull:
		return WhereClause{Query: f.Field + " IS NOT NULL"}
	default: // OpEqual
		return WhereClause{Query: f.Field + " = ?", Args: []any{f.Value}}
	}
}

func parseOperator(s string) (Operator, bool) {
	switch strings.ToLower(s) {
	case "eq":
		return OpEqual, true
	case "neq":
		return OpNotEqual, true
	case "like", "contains":
		return OpContains, true
	case "nlike":
		return OpNotContains, true
	case "in":
		return OpIn, true
	case "nin", "notin":
		return OpNotIn, true
	case "gt":
		return OpGreater, true
	case "gte":
		return OpGreaterEq, true
	case "lt":
		return OpLess, true
	case "lte":
		return OpLessEq, true
	case "null", "isnull":
		return OpIsNull, true
	case "nnull", "notnull":
		return OpIsNotNull, true
	default:
		return "", false
	}
}

// allowSet builds a lookup set; an empty allowlist means "allow nothing" so a
// missing allowlist can never accidentally expose every column.
func allowSet(allowed []string) map[string]bool {
	m := make(map[string]bool, len(allowed))
	for _, f := range allowed {
		m[f] = true
	}
	return m
}

func buildFilter(field, op, value string, hasValue bool) (Filter, bool) {
	o, ok := parseOperator(op)
	if !ok {
		return Filter{}, false
	}
	f := Filter{Field: field, Operator: o}
	switch o {
	case OpIsNull, OpIsNotNull:
		// no value needed
	case OpIn, OpNotIn:
		if !hasValue {
			return Filter{}, false
		}
		f.Value = strings.Split(value, ",")
	default:
		if !hasValue {
			return Filter{}, false
		}
		f.Value = value
	}
	return f, true
}

// ParseSearchQuery parses the search_query DSL — "field.operator.value" terms
// joined by "&" (e.g. "status.eq.success&timestamp.gte.2026-06-01") — into
// filters. Only fields in `allowed` survive; unknown fields/operators are
// dropped. This allowlist is the injection guard.
func ParseSearchQuery(query string, allowed []string) []Filter {
	if query == "" {
		return nil
	}
	allow := allowSet(allowed)
	var filters []Filter
	for _, term := range strings.Split(query, "&") {
		// SplitN with 3 keeps dotted values intact (IPs, RFC3339 timestamps).
		parts := strings.SplitN(term, ".", 3)
		if len(parts) < 2 || !allow[parts[0]] {
			continue
		}
		value := ""
		hasValue := len(parts) == 3
		if hasValue {
			value = parts[2]
		}
		if f, ok := buildFilter(parts[0], parts[1], value, hasValue); ok {
			filters = append(filters, f)
		}
	}
	return filters
}

// ParseQueryParams parses URL query params of the form field.operator=value
// (e.g. ?status.eq=success&priority.in=high,critical) into filters, keeping only
// allowlisted fields.
func ParseQueryParams(params map[string][]string, allowed []string) []Filter {
	allow := allowSet(allowed)
	var filters []Filter
	for key, values := range params {
		if len(values) == 0 {
			continue
		}
		fieldOp := strings.SplitN(key, ".", 2)
		if len(fieldOp) != 2 || !allow[fieldOp[0]] {
			continue
		}
		if f, ok := buildFilter(fieldOp[0], fieldOp[1], values[0], values[0] != ""); ok {
			filters = append(filters, f)
		}
	}
	return filters
}

// ParseFilters parses a "field.operator=value&..." query string into filters,
// keeping only allowlisted fields.
func ParseFilters(query string, allowed []string) []Filter {
	if query == "" {
		return nil
	}
	allow := allowSet(allowed)
	var filters []Filter
	for _, part := range strings.Split(query, "&") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		fieldOp := strings.SplitN(kv[0], ".", 2)
		if len(fieldOp) != 2 || !allow[fieldOp[0]] {
			continue
		}
		if f, ok := buildFilter(fieldOp[0], fieldOp[1], kv[1], kv[1] != ""); ok {
			filters = append(filters, f)
		}
	}
	return filters
}

// ParseSort parses the sort_by DSL — "field.dir" terms joined by "&"
// (e.g. "timestamp.desc&user_login.asc") — into a safe ORDER BY. Only allowlisted
// fields are kept; direction defaults to ASC.
func ParseSort(sort string, allowed []string) string {
	if sort == "" {
		return ""
	}
	allow := allowSet(allowed)
	var parts []string
	for _, term := range strings.Split(sort, "&") {
		kv := strings.SplitN(term, ".", 2)
		if !allow[kv[0]] {
			continue
		}
		dir := "ASC"
		if len(kv) == 2 && strings.EqualFold(kv[1], "desc") {
			dir = "DESC"
		}
		parts = append(parts, kv[0]+" "+dir)
	}
	return strings.Join(parts, ", ")
}

// WithFilters appends parsed filters as WHERE clauses.
func WithFilters(filters []Filter) QueryOption {
	return func(o *QueryOptions) {
		for _, f := range filters {
			o.Where = append(o.Where, f.ToWhereClause())
		}
	}
}

// FilterBuilder is a fluent builder for filters used in server-side code (where
// fields are trusted, not client input).
type FilterBuilder struct {
	filters []Filter
}

func NewFilterBuilder() *FilterBuilder { return &FilterBuilder{} }

func (b *FilterBuilder) Equal(field string, value any) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpEqual, Value: value})
	return b
}

func (b *FilterBuilder) NotEqual(field string, value any) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpNotEqual, Value: value})
	return b
}

func (b *FilterBuilder) Contains(field string, value string) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpContains, Value: value})
	return b
}

func (b *FilterBuilder) In(field string, values any) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpIn, Value: values})
	return b
}

func (b *FilterBuilder) NotIn(field string, values any) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpNotIn, Value: values})
	return b
}

func (b *FilterBuilder) GreaterThan(field string, value any) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpGreater, Value: value})
	return b
}

func (b *FilterBuilder) LessThan(field string, value any) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpLess, Value: value})
	return b
}

func (b *FilterBuilder) IsNull(field string) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpIsNull})
	return b
}

func (b *FilterBuilder) IsNotNull(field string) *FilterBuilder {
	b.filters = append(b.filters, Filter{Field: field, Operator: OpIsNotNull})
	return b
}

// Build returns the accumulated filters as a QueryOption.
func (b *FilterBuilder) Build() QueryOption { return WithFilters(b.filters) }

// Filters returns the raw filter slice.
func (b *FilterBuilder) Filters() []Filter { return b.filters }
