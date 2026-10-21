package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/executor"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

// Dispatcher walks the DAG: it pulls PENDING executions, hands each one to its
// registered Executor, then fires downstream children — one execution per
// fired on_success/on_error edge. Unfired branches are recorded as DEAD rows
// for visibility but never run.
type Dispatcher struct {
	exec  connectors.ExecutionRepository
	runs  connectors.FlowRunRepository
	flows *FlowStore
	vars  connectors.VariableUsecase
	reg   executor.Registry
	kick  chan struct{}
}

func NewDispatcher(
	exec connectors.ExecutionRepository,
	runs connectors.FlowRunRepository,
	flows *FlowStore,
	vars connectors.VariableUsecase,
	reg executor.Registry,
) *Dispatcher {
	return &Dispatcher{exec: exec, runs: runs, flows: flows, vars: vars, reg: reg, kick: make(chan struct{}, 1)}
}

const (
	dispatchTick        = 15 * time.Second
	dispatchBatch       = database.MaxPageSize
	dispatchConcurrency = 5
	dispatchTimeout     = 5 * time.Minute // llm_enrich/llm_action streams can take minutes
	dispatchMaxRetries  = 3
	reapGrace           = 1 * time.Minute
)

var errSettleLostToReaper = errors.New("soar dispatch: settle lost to reaper")

func (d *Dispatcher) Kick() {
	select {
	case d.kick <- struct{}{}:
	default:
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	if len(d.reg) == 0 {
		_ = catcher.Error("soar dispatcher disabled: no executors registered", nil, nil)
		return
	}
	t := time.NewTicker(dispatchTick)
	defer t.Stop()
	d.drain(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			d.drain(ctx)
		case <-d.kick:
			d.drain(ctx)
		}
	}
}

func (d *Dispatcher) drain(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in soar dispatcher drain", nil, map[string]any{"panic": r})
		}
	}()

	d.reapStuck(ctx)

	listCtx, cancel := context.WithTimeout(tenancy.WithAllTenants(ctx), 20*time.Second)
	pending, _, err := d.exec.List(listCtx, connectors.ExecutionFilters{
		Status: domain.ExecutionStatusPending,
		Params: database.Params{Size: dispatchBatch},
	})
	cancel()
	if err != nil {
		_ = catcher.Error("soar dispatcher: failed to list pending executions", err, nil)
		return
	}

	if len(pending) == 0 {
		return
	}

	sem := make(chan struct{}, dispatchConcurrency)
	var wg sync.WaitGroup
	for i := range pending {
		exec := pending[i]
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			d.process(ctx, exec)
		}()
	}
	wg.Wait()
}

// process is the per-execution pipeline: claim → lookup executor → run →
// persist result/output → spawn children (both fired and dead branches) →
// maybe complete the flow run.
func (d *Dispatcher) process(parent context.Context, exec domain.SoarExecution) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in soar dispatch", nil, map[string]any{"panic": r, "execution": exec.ID})
		}
	}()

	ctx, cancel := context.WithTimeout(authz.WithTenantID(parent, exec.TenantID.String()), dispatchTimeout)
	defer cancel()

	claimed, err := d.exec.ClaimPending(ctx, exec.ID, dispatchTimeout)
	if err != nil {
		_ = catcher.Error("soar dispatch: claim failed", err, map[string]any{"execution": exec.ID})
		return
	}
	if !claimed {
		return
	}

	// Manual executions have no flow-run; they still route through the shell
	// executor but skip DAG spawning.
	if exec.Origin == domain.ExecutionOriginManual {
		d.runOnce(ctx, &exec)
		return
	}

	flow := d.loadFlow(exec)
	if flow == nil {
		d.fail(ctx, exec.ID, domain.NonExecutionCauseUnknown)
		return
	}

	// Transition to EXECUTING for visibility while the executor runs.
	executing := domain.ExecutionStatusExecuting
	_ = d.exec.UpdateStatus(ctx, exec.ID, connectors.ExecutionStatusUpdate{Status: &executing})

	// dispatched_at is internal evidence, written once before the side effect
	// goes out: if the worker dies between here and settle, the row proves the
	// command was handed to its executor.
	now := time.Now().UTC()
	_ = d.exec.UpdateStatus(ctx, exec.ID, connectors.ExecutionStatusUpdate{DispatchedAt: &now})

	branch, output := d.invoke(ctx, &exec)
	rest, cancelRest := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancelRest()
	if err := d.settle(rest, &exec, branch, output); err != nil {
		if errors.Is(err, errSettleLostToReaper) {
			catcher.Info("soar dispatch: worker yielded to reaper after settle race", map[string]any{"execution": exec.ID})
			return
		}
		return
	}

	node, ok := flow.Nodes[exec.NodeID]
	if !ok {
		_ = catcher.Error("soar dispatch: node vanished mid-run", nil, map[string]any{"execution": exec.ID, "node": exec.NodeID})
		return
	}
	fired, dead := d.edgesForBranch(node, branch)
	d.fireChildren(rest, flow, exec, fired, branch)
	d.markDeadBranches(rest, flow, exec, dead, branch)

	if exec.FlowRunID != nil {
		if _, err := d.runs.MaybeComplete(rest, *exec.FlowRunID); err != nil {
			_ = catcher.Error("soar dispatch: maybeComplete failed", err, map[string]any{"flowRun": *exec.FlowRunID})
		}
	}
}

// runOnce handles the legacy manual path: run the shell executor once, no DAG.
func (d *Dispatcher) runOnce(ctx context.Context, exec *domain.SoarExecution) error {
	branch, output := d.invoke(ctx, exec)
	return d.settle(ctx, exec, branch, output)
}

func (d *Dispatcher) invoke(ctx context.Context, exec *domain.SoarExecution) (domain.EdgeBranch, json.RawMessage) {
	e, err := d.reg.Lookup(exec.Executor)
	if err != nil {
		exec.Result = err.Error()
		return domain.EdgeBranchError, nil
	}
	output, err := e.Execute(ctx, exec)
	if err != nil {
		if errors.Is(err, executor.ErrAgentOffline) {
			d.handleOffline(ctx, *exec)
			return domain.EdgeBranchError, nil
		}
		if exec.Result == "" {
			exec.Result = err.Error()
		}
		return domain.EdgeBranchError, nil
	}
	return domain.EdgeBranchSuccess, output
}

func (d *Dispatcher) settle(ctx context.Context, exec *domain.SoarExecution, branch domain.EdgeBranch, output json.RawMessage) error {
	result := exec.Result
	if d.vars != nil && result != "" {
		masked, merr := d.vars.MaskSecrets(ctx, result)
		if merr == nil {
			result = masked
		}
	}
	now := time.Now().UTC()
	status := domain.ExecutionStatusExecuted
	if branch == domain.EdgeBranchError {
		status = domain.ExecutionStatusFailed
	}
	applied, err := d.exec.SettleExecuting(ctx, exec.ID, connectors.ExecutionStatusUpdate{
		Status:     &status,
		Result:     &result,
		FinishedAt: &now,
	})
	if err != nil {
		_ = catcher.Error("soar dispatch: settle status write failed", err, map[string]any{"execution": exec.ID, "status": status})
		return err
	}
	if !applied {
		catcher.Info("soar dispatch: settle lost race to reaper, yielding", map[string]any{"execution": exec.ID})
		return errSettleLostToReaper
	}
	exec.Status = status
	if branch == domain.EdgeBranchSuccess && exec.Kind == domain.NodeKindEnrichment && len(output) > 0 {
		if err := d.exec.SaveOutput(ctx, exec.ID, output); err != nil {
			_ = catcher.Error("soar dispatch: failed to save enrichment output", err, map[string]any{"execution": exec.ID})
		}
		exec.Output = output
	}
	return nil
}

var retriableExecutors = map[string]bool{
	"conditional": true,
	"llm_enrich":  true,
	"llm_action":  true,
}

func (d *Dispatcher) reapStuck(ctx context.Context) {
	listCtx, cancel := context.WithTimeout(tenancy.WithAllTenants(ctx), 20*time.Second)
	stuck, err := d.exec.ListStuckExecuting(listCtx, time.Now().UTC().Add(-dispatchTimeout-reapGrace))
	cancel()
	if err != nil {
		_ = catcher.Error("soar dispatcher: failed to list stuck executions", err, nil)
		return
	}
	for i := range stuck {
		d.reapOne(ctx, stuck[i])
	}
}

func (d *Dispatcher) reapOne(ctx context.Context, exec domain.SoarExecution) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in soar reaper", nil, map[string]any{"panic": r, "execution": exec.ID})
		}
	}()

	ctx, cancel := context.WithTimeout(authz.WithTenantID(context.WithoutCancel(ctx), exec.TenantID.String()), 30*time.Second)
	defer cancel()

	if retriableExecutors[exec.Executor] {
		if err := d.exec.RequeueStuck(ctx, exec.ID); err != nil {
			_ = catcher.Error("soar dispatcher: requeue stuck execution failed", err, map[string]any{"execution": exec.ID})
			return
		}
		catcher.Info("soar dispatcher: requeued retriable stuck execution", map[string]any{"execution": exec.ID, "executor": exec.Executor})
		return
	}

	failed := domain.ExecutionStatusFailed
	applied, err := d.exec.CloseStuckLost(ctx, exec.ID)
	if err != nil {
		_ = catcher.Error("soar dispatcher: stuck terminal write failed", err, map[string]any{"execution": exec.ID})
		return
	}
	if !applied {
		return
	}
	exec.Status = failed
	catcher.Info("soar dispatcher: marked stuck execution failed (worker lost)", map[string]any{"execution": exec.ID, "dispatchedAt": exec.DispatchedAt})

	if exec.FlowRunID == nil {
		return
	}
	flow := d.loadFlow(exec)
	if flow == nil {
		_ = catcher.Error("soar reaper: flow vanished for stuck execution", nil, map[string]any{"execution": exec.ID, "rulePath": exec.RulePath})
		d.finishRun(ctx, *exec.FlowRunID)
		return
	}
	node, ok := flow.Nodes[exec.NodeID]
	if !ok {
		_ = catcher.Error("soar reaper: node vanished for stuck execution", nil, map[string]any{"execution": exec.ID, "node": exec.NodeID})
		d.finishRun(ctx, *exec.FlowRunID)
		return
	}
	fired, dead := d.edgesForBranch(node, domain.EdgeBranchError)
	d.markDeadBranches(ctx, flow, exec, fired, domain.EdgeBranchError)
	d.markDeadBranches(ctx, flow, exec, dead, oppositeBranch(domain.EdgeBranchError))
	d.finishRun(ctx, *exec.FlowRunID)
}

func (d *Dispatcher) finishRun(ctx context.Context, runID uuid.UUID) {
	if _, err := d.runs.MaybeComplete(ctx, runID); err != nil {
		_ = catcher.Error("soar reaper: maybeComplete failed", err, map[string]any{"flowRun": runID})
	}
}

func (d *Dispatcher) edgesForBranch(node domain.FlowNode, branch domain.EdgeBranch) (fired []string, dead []string) {
	if branch == domain.EdgeBranchSuccess {
		return node.OnSuccess, node.OnError
	}
	return node.OnError, node.OnSuccess
}

func oppositeBranch(b domain.EdgeBranch) domain.EdgeBranch {
	if b == domain.EdgeBranchSuccess {
		return domain.EdgeBranchError
	}
	return domain.EdgeBranchSuccess
}

func (d *Dispatcher) fireChildren(ctx context.Context, flow *domain.Flow, parent domain.SoarExecution, childIDs []string, branch domain.EdgeBranch) {
	if len(childIDs) == 0 || parent.FlowRunID == nil {
		return
	}
	maxDepth := flow.ResolvedMaxDepth()
	for _, id := range childIDs {
		node, ok := flow.Nodes[id]
		if !ok {
			_ = catcher.Error("soar dispatch: edge points to missing node", nil, map[string]any{"parent": parent.ID, "child": id})
			continue
		}
		childDepth := parent.Depth + 1
		if childDepth > maxDepth {
			_ = catcher.Error("soar dispatch: max_depth exceeded", nil, map[string]any{"parent": parent.ID, "child": id, "depth": childDepth})
			continue
		}
		child, err := d.exec.RecordEdge(ctx, connectors.RecordEdgeRequest{
			FlowRunID:     *parent.FlowRunID,
			TenantID:      parent.TenantID,
			RulePath:      parent.RulePath,
			AlertID:       parent.AlertID,
			Parent:        parent,
			ChildNodeID:   id,
			ChildDepth:    childDepth,
			ChildKind:     node.Kind,
			ChildExecutor: node.Executor,
			Branch:        branch,
			Fired:         true,
		})
		if err != nil {
			_ = catcher.Error("soar dispatch: record edge failed", err, map[string]any{"parent": parent.ID, "child": id})
			continue
		}
		d.transitionChild(ctx, node, child)
	}
}

func (d *Dispatcher) markDeadBranches(ctx context.Context, flow *domain.Flow, parent domain.SoarExecution, childIDs []string, branch domain.EdgeBranch) {
	if len(childIDs) == 0 || parent.FlowRunID == nil {
		return
	}
	maxDepth := flow.ResolvedMaxDepth()
	for _, id := range childIDs {
		node, ok := flow.Nodes[id]
		if !ok {
			continue
		}
		childDepth := parent.Depth + 1
		if childDepth > maxDepth {
			continue
		}
		if err := d.exec.RecordDeadBranch(ctx, connectors.RecordEdgeRequest{
			FlowRunID:     *parent.FlowRunID,
			TenantID:      parent.TenantID,
			RulePath:      parent.RulePath,
			AlertID:       parent.AlertID,
			Parent:        parent,
			ChildNodeID:   id,
			ChildDepth:    childDepth,
			ChildKind:     node.Kind,
			ChildExecutor: node.Executor,
			Branch:        branch,
			Fired:         false,
		}); err != nil {
			_ = catcher.Error("soar dispatch: record dead branch failed", err, map[string]any{"parent": parent.ID, "child": id})
		}
	}
}

func (d *Dispatcher) transitionChild(ctx context.Context, node domain.FlowNode, child *domain.SoarExecution) {
	contribs := []ParentContribution{{Context: child.Context}}
	if child.Kind == domain.NodeKindEnrichment && len(child.Output) > 0 {
		contribs = append(contribs, ParentContribution{EnrichmentNodeID: child.NodeID, Output: child.Output})
	}
	bag := MergeContexts(contribs)

	command, err := Interpolate(ctx, d.vars, bag, node.Command, node.Shell)
	if err != nil {
		_ = catcher.Error("soar dispatch: command interpolation failed", err, map[string]any{"execution": child.ID})
		return
	}
	params, err := InterpolateJSON(ctx, d.vars, bag, node.Params)
	if err != nil {
		_ = catcher.Error("soar dispatch: params interpolation failed", err, map[string]any{"execution": child.ID})
		return
	}
	shell, err := Interpolate(ctx, d.vars, bag, node.Shell, "")
	if err != nil {
		_ = catcher.Error("soar dispatch: shell interpolation failed", err, map[string]any{"execution": child.ID})
		return
	}
	agent := node.Agent
	if agent == "" && child.Agent != "" {
		agent = child.Agent
	}
	if agent != "" {
		if resolved, err := Interpolate(ctx, d.vars, bag, agent, ""); err == nil {
			agent = resolved
		}
	}

	_ = d.exec.TransitionReady(ctx, child.ID, connectors.ReadyUpdate{
		Status:  domain.ExecutionStatusPending,
		Context: bag,
		Params:  params,
		Command: command,
		Shell:   shell,
		Agent:   agent,
	})
	d.Kick()
}

func (d *Dispatcher) loadFlow(exec domain.SoarExecution) *domain.Flow {
	sf := d.flows.Get(exec.TenantID.String(), exec.RulePath)
	if sf == nil {
		return nil
	}
	f := sf.Flow
	return &f
}

func (d *Dispatcher) handleOffline(ctx context.Context, exec domain.SoarExecution) {
	failed := domain.ExecutionStatusFailed
	cause := domain.NonExecutionCauseAgentOffline
	if exec.Retries+1 >= dispatchMaxRetries {
		if _, err := d.exec.SettleExecuting(ctx, exec.ID, connectors.ExecutionStatusUpdate{
			Status:            &failed,
			NonExecutionCause: &cause,
		}); err != nil {
			_ = catcher.Error("soar dispatch: offline terminal write failed", err, map[string]any{"execution": exec.ID})
		}
		return
	}
	if err := d.exec.UpdateStatus(ctx, exec.ID, connectors.ExecutionStatusUpdate{
		NonExecutionCause: &cause,
		IncrementRetries:  true,
	}); err != nil {
		_ = catcher.Error("soar dispatch: offline retry write failed", err, map[string]any{"execution": exec.ID})
	}
}

func (d *Dispatcher) fail(ctx context.Context, id uuid.UUID, cause domain.NonExecutionCause) {
	failed := domain.ExecutionStatusFailed
	if err := d.exec.UpdateStatus(ctx, id, connectors.ExecutionStatusUpdate{
		Status:            &failed,
		NonExecutionCause: &cause,
	}); err != nil {
		_ = catcher.Error("soar dispatch: fail status write failed", err, map[string]any{"execution": id, "cause": cause})
	}
}
