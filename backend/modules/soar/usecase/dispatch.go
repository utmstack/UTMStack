package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type Dispatcher struct {
	repo  connectors.ExecutionRepository
	flows *FlowStore
	agent *agentmanager.AgentManagerClient
	vars  connectors.VariableUsecase
	kick  chan struct{}
}

func NewDispatcher(repo connectors.ExecutionRepository, flows *FlowStore, agent *agentmanager.AgentManagerClient, vars connectors.VariableUsecase) *Dispatcher {
	return &Dispatcher{repo: repo, flows: flows, agent: agent, vars: vars, kick: make(chan struct{}, 1)}
}

const (
	dispatchTick        = 15 * time.Second
	dispatchBatch       = database.MaxPageSize
	dispatchConcurrency = 5
	dispatchTimeout     = 60 * time.Second
	dispatchMaxRetries  = 3
	dispatchOriginType  = "INCIDENT_RESPONSE_AUTOMATION"
	dispatchExecutedBy  = "SYSTEM"
)

func (d *Dispatcher) Kick() {
	select {
	case d.kick <- struct{}{}:
	default:
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	if d.agent == nil {
		_ = catcher.Error("soar dispatcher disabled: no agent-manager client", nil, nil)
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

	listCtx, cancel := context.WithTimeout(tenancy.WithAllTenants(ctx), 20*time.Second)
	pending, _, err := d.repo.List(listCtx, connectors.ExecutionFilters{
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

func (d *Dispatcher) process(parent context.Context, exec domain.SoarExecution) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in soar dispatch", nil, map[string]any{"panic": r, "execution": exec.ID})
		}
	}()

	ctx, cancel := context.WithTimeout(authz.WithTenantID(parent, exec.TenantID.String()), dispatchTimeout)
	defer cancel()

	claimed, err := d.repo.ClaimPending(ctx, exec.ID, dispatchTimeout)
	if err != nil {
		_ = catcher.Error("soar dispatch: claim failed", err, map[string]any{"execution": exec.ID})
		return
	}
	if !claimed {
		return
	}

	flow := d.flows.Get(exec.TenantID.String(), exec.RulePath)
	if flow == nil {
		d.fail(ctx, exec.ID, domain.NonExecutionCauseUnknown)
		return
	}

	agentID, found, err := d.resolveAgent(ctx, exec.Agent)
	if err != nil {
		// Transient lookup failure — leave PENDING for the next tick.
		_ = catcher.Error("soar dispatch: agent lookup failed", err, map[string]any{"execution": exec.ID, "agent": exec.Agent})
		return
	}
	if !found {
		d.fail(ctx, exec.ID, domain.NonExecutionCauseAgentNotFound)
		return
	}

	command := exec.Command
	if d.vars != nil {
		interpolated, ierr := d.vars.InterpolateCommand(ctx, exec.Command)
		if ierr != nil {
			_ = catcher.Error("soar dispatch: variable interpolation failed", ierr, map[string]any{"execution": exec.ID})
		}
		command = interpolated
	}

	cmd := &agent.UtmCommand{
		AgentId:    agentID,
		Command:    command,
		ExecutedBy: dispatchExecutedBy,
		OriginType: dispatchOriginType,
		OriginId:   exec.RulePath,
		Reason:     fmt.Sprintf("Incident response automation: rule %s matched alert %s", exec.RulePath, exec.AlertID),
		Shell:      flow.Shell,
	}

	res, err := d.agent.ProcessCommand(ctx, cmd)
	if err != nil {
		if isOfflineError(err) {
			d.handleOffline(ctx, exec)
			return
		}
		_ = catcher.Error("soar dispatch: command failed", err, map[string]any{"execution": exec.ID})
		d.fail(ctx, exec.ID, domain.NonExecutionCauseUnknown)
		return
	}

	executed := domain.ExecutionStatusExecuted
	result := res.GetResult()

	if d.vars != nil {
		masked, merr := d.vars.MaskSecrets(ctx, result)
		if merr != nil {
			_ = catcher.Error("soar dispatch: mask secrets failed", merr, map[string]any{"execution": exec.ID})
		}
		result = masked
	}
	if err := d.repo.UpdateStatus(ctx, exec.ID, connectors.ExecutionStatusUpdate{
		Status: &executed,
		Result: &result,
	}); err != nil {
		_ = catcher.Error("soar dispatch: failed to persist result", err, map[string]any{"execution": exec.ID})
	}
}

func (d *Dispatcher) handleOffline(ctx context.Context, exec domain.SoarExecution) {
	cause := domain.NonExecutionCauseAgentOffline
	if exec.Retries+1 >= dispatchMaxRetries {
		failed := domain.ExecutionStatusFailed
		_ = d.repo.UpdateStatus(ctx, exec.ID, connectors.ExecutionStatusUpdate{
			Status:            &failed,
			NonExecutionCause: &cause,
		})
		return
	}
	_ = d.repo.UpdateStatus(ctx, exec.ID, connectors.ExecutionStatusUpdate{
		NonExecutionCause: &cause,
		IncrementRetries:  true,
	})
}

func (d *Dispatcher) fail(ctx context.Context, id uuid.UUID, cause domain.NonExecutionCause) {
	failed := domain.ExecutionStatusFailed
	_ = d.repo.UpdateStatus(ctx, id, connectors.ExecutionStatusUpdate{
		Status:            &failed,
		NonExecutionCause: &cause,
	})
}

func (d *Dispatcher) resolveAgent(ctx context.Context, hostname string) (string, bool, error) {
	if hostname == "" {
		return "", false, nil
	}
	rows, _, err := d.agent.ListAgents(ctx, "hostname.Is="+hostname)
	if err != nil {
		return "", false, err
	}
	if len(rows) == 0 {
		return "", false, nil
	}
	return strconv.FormatUint(uint64(rows[0].GetId()), 10), true, nil
}

func isOfflineError(err error) bool {
	if err == nil {
		return false
	}
	if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
		return true
	}
	return strings.Contains(err.Error(), "not found or is disconnected")
}
