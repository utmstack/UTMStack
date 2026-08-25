package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/pkg/authz"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/tidwall/gjson"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type executionUsecase struct {
	repo   connectors.ExecutionRepository
	runs   connectors.FlowRunRepository
	flows  *FlowStore
	agents connectors.AgentUsecase
	vars   connectors.VariableUsecase
	notify func() // signals the dispatcher to drain immediately after enqueue (may be nil)
}

func NewExecutionUsecase(
	repo connectors.ExecutionRepository,
	runs connectors.FlowRunRepository,
	flows *FlowStore,
	agents connectors.AgentUsecase,
	vars connectors.VariableUsecase,
	notify func(),
) connectors.ExecutionUsecase {
	return &executionUsecase{repo: repo, runs: runs, flows: flows, agents: agents, vars: vars, notify: notify}
}

// HandleMatch starts a new flow run — creates the SoarFlowRun row plus one
// PENDING SoarExecution per declared root. Non-root nodes are spawned lazily
// by the dispatcher as their parents complete.
func (u *executionUsecase) HandleMatch(ctx context.Context, req dto.MatchRequest) error {
	tenant := authz.TenantIDFromContext(ctx)
	sf := u.flows.Get(tenant, req.RulePath)
	if sf == nil || !sf.Active() {
		return nil
	}
	flow := sf.Flow
	if len(flow.Roots) == 0 || len(flow.Nodes) == 0 {
		return nil
	}
	alertJSON := req.Alert
	alertID := gjson.GetBytes(alertJSON, "id").String()
	tenantUUID, tenantErr := uuid.Parse(tenant)
	if tenantErr != nil {
		return catcher.Error("soar: invalid tenant id in context", tenantErr, map[string]any{"tenant": tenant})
	}

	run, err := u.runs.Create(ctx, &domain.SoarFlowRun{
		TenantID:  tenantUUID,
		RulePath:  req.RulePath,
		AlertID:   alertID,
		AlertJSON: alertJSON,
		MaxDepth:  flow.ResolvedMaxDepth(),
		Status:    domain.ExecutionStatusPending,
		StartedAt: time.Now().UTC(),
	})
	if err != nil {
		return catcher.Error("soar: failed to create flow run", err, map[string]any{"rule": req.RulePath})
	}

	bag := NewRootContext(alertJSON)

	for _, rootID := range flow.Roots {
		node, ok := flow.Nodes[rootID]
		if !ok {
			_ = catcher.Error("soar: root references missing node", nil, map[string]any{"rule": req.RulePath, "root": rootID})
			continue
		}
		agent, resErr := u.resolveAgentForNode(ctx, flow, node, bag, alertJSON)
		if resErr != nil {
			_ = catcher.Error("soar: agent resolution failed", resErr, map[string]any{"rule": req.RulePath, "root": rootID})
			continue
		}
		params, ierr := InterpolateJSON(ctx, u.vars, bag, node.Params)
		if ierr != nil {
			_ = catcher.Error("soar: params interpolation failed", ierr, map[string]any{"rule": req.RulePath, "root": rootID})
			continue
		}
		command, ierr := Interpolate(ctx, u.vars, bag, node.Command)
		if ierr != nil {
			_ = catcher.Error("soar: command interpolation failed", ierr, map[string]any{"rule": req.RulePath, "root": rootID})
			continue
		}
		exec := &domain.SoarExecution{
			TenantID:  tenantUUID,
			Origin:    domain.ExecutionOriginFlow,
			FlowRunID: &run.ID,
			RulePath:  req.RulePath,
			AlertID:   alertID,
			NodeID:    rootID,
			Depth:     0,
			Kind:      node.Kind,
			Executor:  node.Executor,
			Params:    params,
			Context:   json.RawMessage(bag),
			Command:   command,
			Shell:     node.Shell,
			Agent:     agent,
			Status:    domain.ExecutionStatusPending,
			StartedAt: time.Now().UTC(),
		}
		if _, err := u.repo.Create(ctx, exec); err != nil {
			return catcher.Error("soar: failed to enqueue root execution", err,
				map[string]any{"rule": req.RulePath, "root": rootID})
		}
	}
	if u.notify != nil {
		u.notify()
	}
	return nil
}

// resolveAgentForNode picks the target agent for a shell node. Non-shell
// executors return "" (no agent needed). Node.Agent wins when set; otherwise
// we default to the box that raised the alert (alert.dataSource) — the common
// "restart the service on the affected host" pattern. ExcludedAgents only
// applies in the auto-resolve path: an explicit Agent is treated as an
// operator override that bypasses the deny list.
func (u *executionUsecase) resolveAgentForNode(ctx context.Context, _ domain.Flow, node domain.FlowNode, bag ContextBag, alertJSON []byte) (string, error) {
	if node.Executor != "shell" {
		return "", nil
	}
	if node.Agent != "" {
		return Interpolate(ctx, u.vars, bag, node.Agent)
	}
	src := gjson.GetBytes(alertJSON, "dataSource").String()
	for _, x := range node.ExcludedAgents {
		if x == src {
			return "", nil // returns empty → shell executor fails with ErrAgentNotFound → on_error fires
		}
	}
	return src, nil
}

func (u *executionUsecase) List(ctx context.Context, f dto.ExecutionFilters) (*database.List[dto.ExecutionResponse], error) {
	executions, total, err := u.repo.List(ctx, connectors.ExecutionFilters{
		Origin:            f.Origin,
		RulePath:          f.RulePath,
		AlertID:           f.AlertID,
		Agent:             f.Agent,
		TriggeredBy:       f.TriggeredBy,
		Status:            f.Status,
		NonExecutionCause: f.NonExecutionCause,
		StartedAtGTE:      f.StartedAtGTE,
		StartedAtLTE:      f.StartedAtLTE,
		Params:            f.Params,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.ExecutionResponse, len(executions))
	for i, e := range executions {
		items[i] = dto.ExecutionResponse{
			ID:                e.ID,
			Origin:            e.Origin,
			RulePath:          e.RulePath,
			AlertID:           e.AlertID,
			TriggeredBy:       e.TriggeredBy,
			Agent:             e.Agent,
			Command:           e.Command,
			Result:            e.Result,
			Status:            e.Status,
			StartedAt:         e.StartedAt,
			FinishedAt:        e.FinishedAt,
			NonExecutionCause: e.NonExecutionCause,
			Retries:           e.Retries,
			NodeID:            e.NodeID,
			Kind:              e.Kind,
			Executor:          e.Executor,
			FlowRunID:         e.FlowRunID,
			Depth:             e.Depth,
		}
	}

	return &database.List[dto.ExecutionResponse]{Items: items, Total: total}, nil
}

func (u *executionUsecase) StartManual(ctx context.Context, agent, command, triggeredBy string) (uuid.UUID, error) {
	e, err := u.repo.Create(ctx, &domain.SoarExecution{
		Origin:      domain.ExecutionOriginManual,
		TriggeredBy: triggeredBy,
		Agent:       agent,
		Command:     command,
		Executor:    "shell",
		Kind:        domain.NodeKindExecutor,
		NodeID:      "manual",
		Status:      domain.ExecutionStatusPending,
		StartedAt:   time.Now().UTC(),
	})
	if err != nil {
		return uuid.Nil, catcher.Error("soar: failed to record a manual execution", err,
			map[string]any{"agent": agent, "by": triggeredBy})
	}
	return e.ID, nil
}

func (u *executionUsecase) FinishManual(ctx context.Context, id uuid.UUID, status domain.ExecutionStatus, result string) error {
	now := time.Now().UTC()
	if err := u.repo.UpdateStatus(ctx, id, connectors.ExecutionStatusUpdate{
		Status:     &status,
		Result:     &result,
		FinishedAt: &now,
	}); err != nil {
		return catcher.Error("soar: failed to close a manual execution", err, map[string]any{"execution": id})
	}
	return nil
}
