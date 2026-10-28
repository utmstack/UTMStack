package common_models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/store"
)

const timeColumn = "@timestamp"

func SplitTimeBounds(scope store.Scope, filters []FilterType) (store.Scope, []FilterType) {
	rest := make([]FilterType, 0, len(filters))
	for _, f := range filters {
		if f.Field != timeColumn {
			rest = append(rest, f)
			continue
		}
		switch f.Operator {
		case OpIsBetween:
			from, to := boundPair(f.Value)
			if !from.IsZero() {
				scope.From = from
			}
			if !to.IsZero() {
				scope.To = to
			}
		case OpGreater:
			if t := asTime(f.Value); !t.IsZero() {
				scope.From = t
			}
		case OpLessOrEq:
			if t := asTime(f.Value); !t.IsZero() {
				scope.To = t
			}
		default:
			rest = append(rest, f)
		}
	}
	return scope, rest
}

func boundPair(v any) (time.Time, time.Time) {
	pair, ok := v.([]any)
	if !ok || len(pair) != 2 {
		return time.Time{}, time.Time{}
	}
	return asTime(pair[0]), asTime(pair[1])
}

func asTime(v any) time.Time {
	s, ok := v.(string)
	if !ok || s == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.000", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// ErrInvalidField is a field name the event store cannot be asked about.
var ErrInvalidField = errors.New("invalid field name")

const maxFieldLen = 256

// ValidateField checks a field name that came from a caller before it reaches
// the store, which writes it into the query text. A dotted name addresses a JSON
// path, so no segment may be empty. Backticks and backslashes are what quoting
// has to survive and nothing legitimate uses them; control characters have no
// place in a name at all. Refusing them here also turns what would be a store
// error, and a 500, into a 400 the caller can read.
func ValidateField(field string) error {
	if field == "" || len(field) > maxFieldLen {
		return fmt.Errorf("%w: must be 1 to %d characters", ErrInvalidField, maxFieldLen)
	}
	for _, seg := range strings.Split(field, ".") {
		if seg == "" {
			return fmt.Errorf("%w: %q has an empty segment", ErrInvalidField, field)
		}
	}
	for _, r := range field {
		if r == '`' || r == '\\' || r < 0x20 || r == 0x7f {
			return fmt.Errorf("%w: %q contains a character that is not allowed", ErrInvalidField, field)
		}
	}
	return nil
}

func ToStoreFilters(in []FilterType) ([]store.Filter, error) {
	out := make([]store.Filter, 0, len(in))
	for _, f := range in {
		op, ok := storeOps[f.Operator]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrOperatorUnsupported, f.Operator)
		}
		// A text search names no column; the store searches its own.
		if op != store.OpSearch && op != store.OpNotSearch {
			if err := ValidateField(f.Field); err != nil {
				return nil, err
			}
		}
		out = append(out, store.Filter{Field: f.Field, Op: op, Value: f.Value})
	}
	return out, nil
}

var ErrOperatorUnsupported = fmt.Errorf("the event store does not support this operator")

var storeOps = map[FilterOperator]store.Op{
	OpEquals:    store.OpEq,
	OpNotEquals: store.OpNotEq,

	OpIn:         store.OpIn,
	OpInOr:       store.OpIn,
	OpIsOneOf:    store.OpIn,
	OpIsNotOneOf: store.OpNotIn,

	OpGreater:  store.OpGt,
	OpLessOrEq: store.OpLte,

	OpExists: store.OpExists,

	OpContain:        store.OpContains,
	OpDoesNotContain: store.OpNotContains,
	OpNotContains:    store.OpNotContains,

	OpIsBetween:    store.OpBetween,
	OpIsNotBetween: store.OpNotBetween,

	OpNotExists: store.OpNotExists,

	OpStartWith:    store.OpStartsWith,
	OpNotStartWith: store.OpNotStartsWith,
	OpEndsWith:     store.OpEndsWith,
	OpNotEndsWith:  store.OpNotEndsWith,

	OpIsInFields:    store.OpSearch,
	OpIsNotInFields: store.OpNotSearch,

	OpContainOneOf:        store.OpContainsAny,
	OpDoesNotContainOneOf: store.OpNotContainsAny,
}

func Unsupported() []FilterOperator {
	all := []FilterOperator{
		OpEquals, OpNotEquals, OpIn, OpInOr, OpLessOrEq, OpGreater,
		OpExists, OpNotExists, OpNotContains,
		OpContain, OpDoesNotContain, OpContainOneOf, OpDoesNotContainOneOf,
		OpIsOneOf, OpIsNotOneOf, OpIsBetween, OpIsNotBetween,
		OpIsInFields, OpIsNotInFields,
		OpEndsWith, OpNotEndsWith, OpStartWith, OpNotStartWith,
	}
	out := make([]FilterOperator, 0)
	for _, op := range all {
		if _, ok := storeOps[op]; !ok {
			out = append(out, op)
		}
	}
	return out
}
