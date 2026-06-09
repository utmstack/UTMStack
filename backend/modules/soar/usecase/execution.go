package usecase

import (
	"context"
	"regexp"
	"strings"

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

var commandPlaceholderRE = regexp.MustCompile(`\$\((.*?)\)`)

func (u *executionUsecase) HandleMatch(ctx context.Context, req dto.MatchRequest) error {
	sf := u.flows.Get(req.RulePath)
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
	enqueued := false
	for _, raw := range flow.Commands {
		command := buildCommand(raw, alertJSON)
		if _, err := u.repo.Create(ctx, &domain.AlertResponseRuleExecution{
			RulePath:        req.RulePath,
			AlertID:         alertID,
			Command:         command,
			Agent:           target,
			ExecutionStatus: domain.ExecutionStatusPending,
		}); err != nil {
			_ = catcher.Error("soar: failed to enqueue execution", err, map[string]any{"rule": req.RulePath, "alert": alertID})
			continue
		}
		enqueued = true
	}
	if enqueued && u.notify != nil {
		u.notify()
	}
	return nil
}

func (u *executionUsecase) resolveAgent(ctx context.Context, flow Flow, alertJSON string) (string, error) {
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

// buildCommand substitutes $(field.path) placeholders with values from the alert.
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
		ID:                f.ID,
		RulePath:          f.RulePath,
		AlertID:           f.AlertID,
		Agent:             f.Agent,
		ExecutionStatus:   f.ExecutionStatus,
		NonExecutionCause: f.NonExecutionCause,
		ExecutionDateGTE:  f.ExecutionDateGTE,
		ExecutionDateLTE:  f.ExecutionDateLTE,
		Params:            f.Params,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.ExecutionResponse, len(executions))
	for i, e := range executions {
		items[i] = dto.ExecutionResponse{
			ID:                e.ID,
			RulePath:          e.RulePath,
			AlertID:           e.AlertID,
			Command:           e.Command,
			CommandResult:     e.CommandResult,
			Agent:             e.Agent,
			ExecutionDate:     e.ExecutionDate,
			ExecutionStatus:   e.ExecutionStatus,
			NonExecutionCause: e.NonExecutionCause,
			ExecutionRetries:  e.ExecutionRetries,
		}
	}

	return &database.List[dto.ExecutionResponse]{Items: items, Total: total}, nil
}
