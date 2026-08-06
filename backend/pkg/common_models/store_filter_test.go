package common_models

import (
	"errors"
	"testing"

	"github.com/threatwinds/go-sdk/store"
)

func TestTheOperatorsThatMapDoMap(t *testing.T) {
	got, err := ToStoreFilters([]FilterType{
		{Field: "a", Operator: OpEquals, Value: "x"},
		{Field: "b", Operator: OpIsBetween, Value: []any{1, 2}},
		{Field: "c", Operator: OpContain, Value: "y"},
	})
	if err != nil {
		t.Fatalf("ToStoreFilters: %v", err)
	}
	want := []store.Op{store.OpEq, store.OpBetween, store.OpContains}
	for i, w := range want {
		if got[i].Op != w {
			t.Errorf("filter %d = %s, want %s", i, got[i].Op, w)
		}
	}
}

// Every operator the DSL defines now maps. A new one added without a
// translation shows up here rather than as an error in front of a user.
func TestEveryOperatorTranslates(t *testing.T) {
	if left := Unsupported(); len(left) != 0 {
		t.Errorf("%d operators do not translate: %v", len(left), left)
	}
}

// One the DSL does not define is still an error rather than being ignored.
func TestAnUnknownOperatorIsAnError(t *testing.T) {
	_, err := ToStoreFilters([]FilterType{{Field: "a", Operator: "REGEX_MATCH", Value: "x"}})
	if !errors.Is(err, ErrOperatorUnsupported) {
		t.Errorf("err = %v, want ErrOperatorUnsupported", err)
	}
}

// A "starts with" is not a "contains": it must not quietly become one.
func TestAnchoredMatchingStaysAnchored(t *testing.T) {
	for op, want := range map[FilterOperator]store.Op{
		OpStartWith:    store.OpStartsWith,
		OpNotStartWith: store.OpNotStartsWith,
		OpEndsWith:     store.OpEndsWith,
		OpNotEndsWith:  store.OpNotEndsWith,
	} {
		if got := storeOps[op]; got != want {
			t.Errorf("%s mapped to %s, want %s", op, got, want)
		}
	}
}
