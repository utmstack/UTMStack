package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// Conditional evaluates a list of predicates against the execution's merged
// context bag and returns success only when all of them are true (AND). A
// failing predicate returns an error so the dispatcher routes the flow down
// the node's onError branch — no bespoke edge kind needed.
// ponytail: reuses domain.FilterType and gjson (already vendored via variable
// + execution interpolation); OnSuccess/OnError already model the true/false
// exits of a conditional.
type Conditional struct{}

func NewConditional() *Conditional { return &Conditional{} }

func (Conditional) Type() string { return "conditional" }

type conditionalParams struct {
	Conditions []domain.FilterType `json:"conditions"`
}

func (c *Conditional) Execute(_ context.Context, exec *domain.SoarExecution) (json.RawMessage, error) {
	var p conditionalParams
	if len(exec.Params) > 0 {
		if err := json.Unmarshal(exec.Params, &p); err != nil {
			return nil, fmt.Errorf("soar conditional: params: %w", err)
		}
	}
	if len(p.Conditions) == 0 {
		return nil, errors.New("soar conditional: at least one condition is required")
	}
	src := string(exec.Context)
	if src == "" {
		src = "{}"
	}
	for _, cond := range p.Conditions {
		if !evaluateFilter(src, cond) {
			exec.Result = fmt.Sprintf("condition failed: %s %s %v", cond.Field, cond.Operator, cond.Value)
			return nil, errors.New(exec.Result)
		}
	}
	exec.Result = "all conditions matched"
	return nil, nil
}

func evaluateFilter(src string, cond domain.FilterType) bool {
	val := gjson.Get(src, cond.Field)
	switch cond.Operator {
	case domain.OperatorExists:
		return val.Exists()
	case domain.OperatorNotExists:
		return !val.Exists()
	}
	got := val.String()
	switch cond.Operator {
	case domain.OperatorIS:
		return got == asString(cond.Value)
	case domain.OperatorISNot:
		return got != asString(cond.Value)
	case domain.OperatorContains:
		return strings.Contains(got, asString(cond.Value))
	case domain.OperatorNotContains:
		return !strings.Contains(got, asString(cond.Value))
	case domain.OperatorStartWith:
		return strings.HasPrefix(got, asString(cond.Value))
	case domain.OperatorNotStartWith:
		return !strings.HasPrefix(got, asString(cond.Value))
	case domain.OperatorEndsWith:
		return strings.HasSuffix(got, asString(cond.Value))
	case domain.OperatorNotEndsWith:
		return !strings.HasSuffix(got, asString(cond.Value))
	case domain.OperatorIsOneOf:
		return oneOf(asStringSlice(cond.Value), got)
	case domain.OperatorIsNotOneOf:
		return !oneOf(asStringSlice(cond.Value), got)
	}
	return false
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func asStringSlice(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			out = append(out, asString(x))
		}
		return out
	}
	return nil
}

func oneOf(hay []string, needle string) bool {
	for _, h := range hay {
		if h == needle {
			return true
		}
	}
	return false
}
