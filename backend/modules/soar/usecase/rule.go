package usecase

import (
	"context"
	"errors"
	"os"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

// ruleUsecase is the SOAR flow (alert-response rule) control plane, backed by
// the file store (YAML-direct) rather than Postgres. Execution history stays in
// the DB and is served by the execution usecase.
type ruleUsecase struct {
	store   *FlowStore
	resolve connectors.ResolveFilterRepository
}

func NewRuleUsecase(store *FlowStore, resolve connectors.ResolveFilterRepository) connectors.RuleUsecase {
	return &ruleUsecase{store: store, resolve: resolve}
}

func (u *ruleUsecase) Create(ctx context.Context, req dto.CreateRuleRequest, createdBy string) (*dto.RuleResponse, error) {
	sf, err := u.store.Create(requestToFlow(req.Name, req.Description, req.Conditions, req.Commands, req.Shell, req.AgentPlatform, req.DefaultAgent, req.ExcludedAgents))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	if req.Active != nil && !*req.Active {
		_ = u.store.SetEnabled(sf.RelPath, false)
		sf = u.store.Get(sf.RelPath)
	}
	return storedFlowToResponse(sf), nil
}

func (u *ruleUsecase) Update(ctx context.Context, relPath string, req dto.UpdateRuleRequest, modifiedBy string) (*dto.RuleResponse, error) {
	sf, err := u.store.Update(relPath, requestToFlow(req.Name, req.Description, req.Conditions, req.Commands, req.Shell, req.AgentPlatform, req.DefaultAgent, req.ExcludedAgents))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	if req.Active != nil {
		_ = u.store.SetEnabled(relPath, *req.Active)
		sf = u.store.Get(relPath)
	}
	return storedFlowToResponse(sf), nil
}

func (u *ruleUsecase) Get(ctx context.Context, relPath string) (*dto.RuleResponse, error) {
	sf := u.store.Get(relPath)
	if sf == nil {
		return nil, domain.ErrRuleNotFound
	}
	return storedFlowToResponse(sf), nil
}

func (u *ruleUsecase) Delete(ctx context.Context, relPath string) error {
	return mapStoreErr(u.store.Delete(relPath))
}

func (u *ruleUsecase) SetEnabled(ctx context.Context, relPath string, enabled bool) error {
	return mapStoreErr(u.store.SetEnabled(relPath, enabled))
}

func (u *ruleUsecase) List(ctx context.Context, f dto.RuleFilters) (*database.List[dto.RuleResponse], error) {
	flows, total := u.store.List(FlowListFilter{
		Page:          f.Page,
		Size:          f.Size,
		Name:          f.RuleName,
		Active:        f.RuleActive,
		SystemOwner:   f.SystemOwner,
		AgentPlatform: f.AgentPlatform,
	})
	items := make([]dto.RuleResponse, 0, len(flows))
	for _, sf := range flows {
		items = append(items, *storedFlowToResponse(sf))
	}
	return &database.List[dto.RuleResponse]{Items: items, Total: int64(total)}, nil
}

func (u *ruleUsecase) ResolveFilterValues(ctx context.Context) (*dto.ResolveFilterValuesResponse, error) {
	platforms, err := u.resolve.GetAgentPlatforms(ctx)
	if err != nil {
		return nil, err
	}
	users, err := u.resolve.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.ResolveFilterValuesResponse{AgentPlatforms: platforms, Users: users}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

// mapStoreErr maps file-store sentinel errors onto the domain errors the handler
// translates into HTTP statuses.
func mapStoreErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, os.ErrExist):
		return domain.ErrRuleNameTaken
	case errors.Is(err, ErrFlowNotFound):
		return domain.ErrRuleNotFound
	case errors.Is(err, ErrSystemFlowContent):
		return domain.ErrSystemRuleReadOnly
	default:
		return err
	}
}

func requestToFlow(name, description string, conds []dto.FilterVM, commands []dto.FlowCommandVM, shell, agentPlatform, defaultAgent string, excludedAgents []string) Flow {
	return Flow{
		Name:           name,
		Description:    description,
		Conditions:     toFlowConditions(conds),
		Commands:       toFlowCommands(commands),
		Shell:          shell,
		AgentPlatform:  agentPlatform,
		DefaultAgent:   defaultAgent,
		ExcludedAgents: excludedAgents,
	}
}

func toFlowConditions(vms []dto.FilterVM) []FlowCondition {
	out := make([]FlowCondition, 0, len(vms))
	for _, v := range vms {
		out = append(out, FlowCondition{Operator: string(v.Operator), Field: v.Field, Value: v.Value})
	}
	return out
}

func toFlowCommands(vms []dto.FlowCommandVM) []FlowCommand {
	out := make([]FlowCommand, 0, len(vms))
	for _, v := range vms {
		out = append(out, FlowCommand{Command: v.Command, Condition: v.Condition})
	}
	return out
}

func flowCommandsToVMs(cmds []FlowCommand) []dto.FlowCommandVM {
	out := make([]dto.FlowCommandVM, 0, len(cmds))
	for _, c := range cmds {
		out = append(out, dto.FlowCommandVM{Command: c.Command, Condition: c.Condition})
	}
	return out
}

func storedFlowToResponse(sf *StoredFlow) *dto.RuleResponse {
	if sf == nil {
		return nil
	}
	resp := &dto.RuleResponse{
		RelPath:        sf.RelPath,
		Name:           sf.Name,
		Description:    sf.Description,
		Commands:       flowCommandsToVMs(sf.Commands),
		Active:         sf.Active(),
		AgentPlatform:  sf.AgentPlatform,
		DefaultAgent:   sf.DefaultAgent,
		Shell:          sf.Shell,
		SystemOwner:    sf.SystemOwned(),
		ExcludedAgents: sf.ExcludedAgents,
	}
	for _, c := range sf.Conditions {
		resp.Conditions = append(resp.Conditions, dto.FilterVM{Operator: domain.OperatorType(c.Operator), Field: c.Field, Value: c.Value})
	}
	if !sf.Modified.IsZero() {
		m := sf.Modified
		resp.LastModifiedDate = &m
	}
	return resp
}
