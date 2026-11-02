package usecase

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type flowRunSpy struct {
	connectors.FlowRunRepository
	run *domain.SoarFlowRun
}

func (r *flowRunSpy) Create(_ context.Context, run *domain.SoarFlowRun) (*domain.SoarFlowRun, error) {
	run.ID = uuid.New()
	r.run = run
	return run, nil
}

type execRecorder struct {
	connectors.ExecutionRepository
	created []*domain.SoarExecution
}

func (r *execRecorder) Create(_ context.Context, e *domain.SoarExecution) (*domain.SoarExecution, error) {
	e.ID = uuid.New()
	r.created = append(r.created, e)
	return e, nil
}

func startFlowStore(t *testing.T, active bool) *FlowStore {
	t.Helper()
	root := t.TempDir()
	store := NewFlowStore(filepath.Join(root, "system"), filepath.Join(root, "user"))
	flow := domain.Flow{
		Name:       "restart",
		Conditions: []domain.FilterType{{Operator: "equals", Field: "id", Value: "a1"}},
		Roots:      []string{"r1"},
		Nodes: map[string]domain.FlowNode{
			"r1": {Kind: domain.NodeKindExecutor, Executor: "shell", Command: "echo hi"},
		},
	}
	if _, err := store.Create(tenantID, flow); err != nil {
		t.Fatal(err)
	}
	if err := store.SetEnabled(tenantID, "restart.yaml", active); err != nil {
		t.Fatal(err)
	}
	return store
}

var tenantID = "8f1c1b8e-0000-4000-8000-000000000001"

type testVars struct{ connectors.VariableUsecase }

func (testVars) MaskSecrets(_ context.Context, s string) (string, error) { return s, nil }
func (testVars) InterpolateCommand(_ context.Context, cmd, shell string) (string, error) {
	return cmd, nil
}

func TestStartFlowEnqueuesRootsAndNotifies(t *testing.T) {
	store := startFlowStore(t, true)
	ex := &execRecorder{}
	runs := &flowRunSpy{}
	notified := false
	uc := &executionUsecase{repo: ex, runs: runs, flows: store, vars: testVars{}, notify: func() { notified = true }}
	ctx := authz.WithTenantID(context.Background(), tenantID)

	_, err := uc.StartFlow(ctx, "restart.yaml", json.RawMessage(`{"id":"a1","dataSource":"host-x"}`), "", "admin@example.com")
	if err != nil {
		t.Fatalf("StartFlow: %v", err)
	}
	if runs.run == nil || runs.run.AlertID != "a1" {
		t.Fatalf("run = %+v, want AlertID a1", runs.run)
	}
	if len(ex.created) != 1 {
		t.Fatalf("enqueued %d roots, want 1", len(ex.created))
	}
	root := ex.created[0]
	if root.Origin != domain.ExecutionOriginFlow || root.Status != domain.ExecutionStatusPending {
		t.Errorf("root origin/status = %s/%s, want FLOW/PENDING (dispatcher must own it)", root.Origin, root.Status)
	}
	if root.FlowRunID == nil || *root.FlowRunID != runs.run.ID {
		t.Error("root not linked to the flow run")
	}
	if root.TriggeredBy != "admin@example.com" {
		t.Errorf("TriggeredBy = %q, want the operator login", root.TriggeredBy)
	}
	if root.Agent != "host-x" {
		t.Errorf("Agent = %q, want dataSource from the alert", root.Agent)
	}
	if !notified {
		t.Error("dispatcher was not kicked after enqueue")
	}
}

// A shell root with no agent in the flow and no dataSource in the alert falls
// back to the operator's choice — the manual trigger exists precisely for
// flows whose natural target is "the box that raised the alert".
func TestStartFlowAgentOverrideCoversMissingDataSource(t *testing.T) {
	store := startFlowStore(t, true)
	ex := &execRecorder{}
	uc := &executionUsecase{repo: ex, runs: &flowRunSpy{}, flows: store}
	ctx := authz.WithTenantID(context.Background(), tenantID)

	_, err := uc.StartFlow(ctx, "restart.yaml", json.RawMessage(`{}`), "host-y", "")
	if err != nil {
		t.Fatalf("StartFlow: %v", err)
	}
	if got := ex.created[0].Agent; got != "host-y" {
		t.Errorf("Agent = %q, want the override %q", got, "host-y")
	}
}

// An empty alert body still runs: the interpolation bag is just {"alert":{}}.
func TestStartFlowToleratesEmptyAlert(t *testing.T) {
	store := startFlowStore(t, true)
	ex := &execRecorder{}
	uc := &executionUsecase{repo: ex, runs: &flowRunSpy{}, flows: store}
	ctx := authz.WithTenantID(context.Background(), tenantID)

	if _, err := uc.StartFlow(ctx, "restart.yaml", nil, "host-z", ""); err != nil {
		t.Fatalf("StartFlow: %v", err)
	}
	if len(ex.created) != 1 {
		t.Fatalf("enqueued %d roots, want 1", len(ex.created))
	}
}

// A flow the tenant switched off must not start from the UI either — same
// guard as the match path, different error so the caller can say "start it
// first" instead of silently dropping the request.
func TestStartFlowRefusesADisabledFlow(t *testing.T) {
	store := startFlowStore(t, false)
	uc := &executionUsecase{repo: &execRecorder{}, runs: &flowRunSpy{}, flows: store}
	ctx := authz.WithTenantID(context.Background(), tenantID)

	_, err := uc.StartFlow(ctx, "restart.yaml", json.RawMessage(`{}`), "", "")
	if err != domain.ErrFlowNotStarted {
		t.Errorf("err = %v, want ErrFlowNotStarted", err)
	}
}
