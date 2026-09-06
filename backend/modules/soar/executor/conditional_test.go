package executor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

func TestConditional_AllMatchTakesSuccessBranch(t *testing.T) {
	c := NewConditional()
	exec := &domain.SoarExecution{
		Kind:    domain.NodeKindExecutor,
		Context: json.RawMessage(`{"alert":{"severity":"high","tags":["prod","edr"]}}`),
		Params: json.RawMessage(`{"conditions":[
			{"field":"alert.severity","operator":"IS","value":"high"},
			{"field":"alert.tags","operator":"CONTAINS","value":"edr"}
		]}`),
	}
	if _, err := c.Execute(context.Background(), exec); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestConditional_MismatchRoutesToOnError(t *testing.T) {
	c := NewConditional()
	exec := &domain.SoarExecution{
		Kind:    domain.NodeKindExecutor,
		Context: json.RawMessage(`{"alert":{"severity":"low"}}`),
		Params:  json.RawMessage(`{"conditions":[{"field":"alert.severity","operator":"IS","value":"high"}]}`),
	}
	if _, err := c.Execute(context.Background(), exec); err == nil {
		t.Fatal("expected error so the dispatcher takes the onError branch")
	}
}

func TestConditional_MissingParamsFails(t *testing.T) {
	c := NewConditional()
	exec := &domain.SoarExecution{Context: json.RawMessage(`{}`)}
	if _, err := c.Execute(context.Background(), exec); err == nil {
		t.Fatal("expected error when no conditions are configured")
	}
}
