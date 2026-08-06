package common_models

import (
	"fmt"

	"github.com/threatwinds/go-sdk/store"
)

func ToStoreFilters(in []FilterType) ([]store.Filter, error) {
	out := make([]store.Filter, 0, len(in))
	for _, f := range in {
		op, ok := storeOps[f.Operator]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrOperatorUnsupported, f.Operator)
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
