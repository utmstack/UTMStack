package usecase

import (
	"context"
	"regexp"
	"strings"

	"github.com/utmstack/utmstack/backend/pkg/authz"

	"time"

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
	flows  *FlowStore
	agents connectors.AgentUsecase
	notify func() // signals the dispatcher to drain immediately after enqueue (may be nil)
}

func NewExecutionUsecase(repo connectors.ExecutionRepository, flows *FlowStore, agents connectors.AgentUsecase, notify func()) connectors.ExecutionUsecase {
	return &executionUsecase{repo: repo, flows: flows, agents: agents, notify: notify}
}

var commandPlaceholderRE = regexp.MustCompile(`\$\(([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z0-9_]+)*)\)`)

func (u *executionUsecase) HandleMatch(ctx context.Context, req dto.MatchRequest) error {
	sf := u.flows.Get(authz.TenantIDFromContext(ctx), req.RulePath)
	if sf == nil || !sf.Active() {
		return nil // flow disabled/deleted since the plugin last loaded it — ignore
	}
	flow := sf.Flow
	alertJSON := string(req.Alert)

	target, err := u.resolveAgent(ctx, flow, alertJSON)
	if err != nil {
		return err
	}
	if target == "" {
		return nil // nothing to run on (source unmanaged + no default, or excluded)
	}

	alertID := gjson.Get(alertJSON, "id").String()
	command := buildCommand(assembleChain(flow.Commands), alertJSON)
	if command == "" {
		return nil
	}
	if _, err := u.repo.Create(ctx, &domain.SoarExecution{
		Origin:    domain.ExecutionOriginFlow,
		RulePath:  req.RulePath,
		AlertID:   alertID,
		Command:   command,
		Agent:     target,
		Status:    domain.ExecutionStatusPending,
		StartedAt: time.Now().UTC(),
	}); err != nil {
		return catcher.Error("soar: failed to enqueue execution", err, map[string]any{"rule": req.RulePath, "alert": alertID})
	}
	if u.notify != nil {
		u.notify()
	}
	return nil
}

func assembleChain(cmds []domain.FlowCommand) string {
	var b strings.Builder
	for i, c := range cmds {
		if c.Command == "" {
			continue
		}
		if b.Len() > 0 {
			op := domain.ConditionAlways.Operator()
			if i > 0 && c.Condition != nil {
				op = c.Condition.Operator()
			}
			b.WriteByte(' ')
			b.WriteString(op)
			b.WriteByte(' ')
		}
		b.WriteString(c.Command)
	}
	return b.String()
}

func (u *executionUsecase) resolveAgent(ctx context.Context, flow domain.Flow, alertJSON string) (string, error) {
	src := gjson.Get(alertJSON, "dataSource").String()
	if agentInList(flow.ExcludedAgents, src) {
		return "", nil
	}
	platformAgents, err := u.agents.ListByPlatform(ctx, flow.AgentPlatform)
	if err != nil {
		return "", err
	}
	if src != "" && agentInList(platformAgents, src) {
		return src, nil
	}
	if flow.DefaultAgent != "" {
		return flow.DefaultAgent, nil
	}
	return "", nil
}

func buildCommand(template, alertJSON string) string {
	return commandPlaceholderRE.ReplaceAllStringFunc(template, func(match string) string {
		field := strings.TrimSuffix(strings.TrimPrefix(match, "$("), ")")
		val := gjson.Get(alertJSON, field)
		if !val.Exists() {
			return match
		}
		return val.String()
	})
}

func agentInList(list []string, v string) bool {
	if v == "" {
		return false
	}
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
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
