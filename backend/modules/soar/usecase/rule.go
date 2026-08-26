package usecase

import (
	"context"
	"errors"
	"os"

	"github.com/utmstack/utmstack/backend/pkg/authz"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type ruleUsecase struct {
	store   *FlowStore
	resolve connectors.ResolveFilterRepository
}

func NewRuleUsecase(store *FlowStore, resolve connectors.ResolveFilterRepository) connectors.RuleUsecase {
	return &ruleUsecase{store: store, resolve: resolve}
}

func (u *ruleUsecase) Create(ctx context.Context, req dto.CreateRuleRequest, createdBy string) (*dto.RuleResponse, error) {
	sf, err := u.store.Create(tenantOf(ctx), requestToFlow(req.Name, req.Description, req.Conditions, req.Roots, req.Nodes, req.MaxDepth))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	if req.Active != nil && !*req.Active {
		_ = u.store.SetEnabled(tenantOf(ctx), sf.RelPath, false)
		sf = u.store.Get(tenantOf(ctx), sf.RelPath)
	}
	return storedFlowToResponse(sf), nil
}

func (u *ruleUsecase) Update(ctx context.Context, relPath string, req dto.UpdateRuleRequest, modifiedBy string) (*dto.RuleResponse, error) {
	sf, err := u.store.Update(tenantOf(ctx), relPath, requestToFlow(req.Name, req.Description, req.Conditions, req.Roots, req.Nodes, req.MaxDepth))
	if err != nil {
		return nil, mapStoreErr(err)
	}
	if req.Active != nil {
		_ = u.store.SetEnabled(tenantOf(ctx), relPath, *req.Active)
		sf = u.store.Get(tenantOf(ctx), relPath)
	}
	return storedFlowToResponse(sf), nil
}

func (u *ruleUsecase) Get(ctx context.Context, relPath string) (*dto.RuleResponse, error) {
	sf := u.store.Get(tenantOf(ctx), relPath)
	if sf == nil {
		return nil, domain.ErrFlowNotFound
	}
	return storedFlowToResponse(sf), nil
}

func (u *ruleUsecase) Delete(ctx context.Context, relPath string) error {
	return mapStoreErr(u.store.Delete(tenantOf(ctx), relPath))
}

func (u *ruleUsecase) SetEnabled(ctx context.Context, relPath string, enabled bool) error {
	return mapStoreErr(u.store.SetEnabled(tenantOf(ctx), relPath, enabled))
}

func (u *ruleUsecase) List(ctx context.Context, f dto.RuleFilters) (*database.List[dto.RuleResponse], error) {
	flows, total := u.store.List(tenantOf(ctx), FlowListFilter{
		Page:        f.Page,
		Size:        f.Size,
		Name:        f.RuleName,
		Active:      f.RuleActive,
		SystemOwner: f.SystemOwner,
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

func mapStoreErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, os.ErrExist):
		return domain.ErrRuleNameTaken
	case errors.Is(err, domain.ErrFlowNotFound):
		return domain.ErrFlowNotFound
	case errors.Is(err, domain.ErrSystemFlowContent):
		return domain.ErrSystemRuleReadOnly
	default:
		return err
	}
}

func requestToFlow(name, description string, conds []dto.FilterVM, roots []string, nodes map[string]dto.FlowNodeVM, maxDepth int) domain.Flow {
	return domain.Flow{
		Name:        name,
		Description: description,
		Conditions:  toFlowConditions(conds),
		Roots:       roots,
		Nodes:       toFlowNodes(nodes),
		MaxDepth:    maxDepth,
	}
}

func toFlowConditions(vms []dto.FilterVM) []domain.FilterType {
	out := make([]domain.FilterType, 0, len(vms))
	for _, v := range vms {
		out = append(out, domain.FilterType{Operator: v.Operator, Field: v.Field, Value: v.Value})
	}
	return out
}

func toFlowNodes(vms map[string]dto.FlowNodeVM) map[string]domain.FlowNode {
	out := make(map[string]domain.FlowNode, len(vms))
	for id, v := range vms {
		out[id] = domain.FlowNode{
			Kind:           v.Kind,
			Executor:       v.Executor,
			Command:        v.Command,
			Shell:          v.Shell,
			Platform:       v.Platform,
			Agent:          v.Agent,
			ExcludedAgents: v.ExcludedAgents,
			Params:         v.Params,
			OnSuccess:      v.OnSuccess,
			OnError:        v.OnError,
		}
	}
	return out
}

func flowNodesToVMs(nodes map[string]domain.FlowNode) map[string]dto.FlowNodeVM {
	out := make(map[string]dto.FlowNodeVM, len(nodes))
	for id, n := range nodes {
		out[id] = dto.FlowNodeVM{
			Kind:           n.Kind,
			Executor:       n.Executor,
			Command:        n.Command,
			Shell:          n.Shell,
			Platform:       n.Platform,
			Agent:          n.Agent,
			ExcludedAgents: n.ExcludedAgents,
			Params:         n.Params,
			OnSuccess:      n.OnSuccess,
			OnError:        n.OnError,
		}
	}
	return out
}

func storedFlowToResponse(sf *domain.StoredFlow) *dto.RuleResponse {
	if sf == nil {
		return nil
	}
	resp := &dto.RuleResponse{
		RelPath:     sf.RelPath,
		Name:        sf.Name,
		Description: sf.Description,
		Roots:       sf.Roots,
		Nodes:       flowNodesToVMs(sf.Nodes),
		MaxDepth:    sf.MaxDepth,
		Active:      sf.Active(),
		SystemOwner: sf.SystemOwned(),
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

func tenantOf(ctx context.Context) string {
	if t := authz.TenantIDFromContext(ctx); t != "" {
		return t
	}
	return authz.DefaultTenantID
}
