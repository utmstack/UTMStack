package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type ruleUsecase struct {
	rules   connectors.RuleRepository
	resolve connectors.ResolveFilterRepository
}

func NewRuleUsecase(
	rules connectors.RuleRepository,
	resolve connectors.ResolveFilterRepository,
) connectors.RuleUsecase {
	return &ruleUsecase{
		rules:   rules,
		resolve: resolve,
	}
}

func (u *ruleUsecase) Create(ctx context.Context, req dto.CreateRuleRequest, createdBy string) (*dto.RuleResponse, error) {
	conditionsJSON, err := marshalConditions(req.Conditions)
	if err != nil {
		return nil, err
	}

	rule := &domain.AlertResponseRule{
		RuleName:        req.Name,
		RuleDescription: req.Description,
		RuleConditions:  conditionsJSON,
		RuleCmd:         req.Command,
		RuleActive:      derefBool(req.Active),
		AgentPlatform:   req.AgentPlatform,
		DefaultAgent:    req.DefaultAgent,
		RuleShell:       req.Shell,
		ExcludedAgents:  strings.Join(req.ExcludedAgents, ","),
		SystemOwner:     false,
		CreatedBy:       createdBy,
		Templates:       templateDigestsToEntities(req.Actions),
	}

	// Rule insert + template sync are wrapped in a DB transaction inside
	// the repository Create method.
	saved, err := u.rules.Create(ctx, rule)
	if err != nil {
		return nil, err
	}
	return ruleToResponse(saved), nil
}

func (u *ruleUsecase) Update(ctx context.Context, req dto.UpdateRuleRequest, modifiedBy string) (*dto.RuleResponse, error) {
	if req.ID == nil || *req.ID == 0 {
		return nil, domain.ErrIDRequired
	}

	conditionsJSON, err := marshalConditions(req.Conditions)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	rule := &domain.AlertResponseRule{
		ID:               *req.ID,
		RuleName:         req.Name,
		RuleDescription:  req.Description,
		RuleConditions:   conditionsJSON,
		RuleCmd:          req.Command,
		RuleActive:       derefBool(req.Active),
		AgentPlatform:    req.AgentPlatform,
		DefaultAgent:     req.DefaultAgent,
		RuleShell:        req.Shell,
		ExcludedAgents:   strings.Join(req.ExcludedAgents, ","),
		LastModifiedBy:   modifiedBy,
		LastModifiedDate: &now,
		Templates:        templateDigestsToEntities(req.Actions),
	}

	saved, err := u.rules.Update(ctx, rule)
	if err != nil {
		return nil, err
	}
	return ruleToResponse(saved), nil
}

func (u *ruleUsecase) GetByID(ctx context.Context, id int64) (*dto.RuleResponse, error) {
	rule, err := u.rules.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ruleToResponse(rule), nil
}

func (u *ruleUsecase) List(ctx context.Context, f dto.RuleFilters) (*database.List[dto.RuleResponse], error) {
	rules, total, err := u.rules.List(ctx, connectors.RuleFilters{
		ID:                  f.ID,
		RuleName:            f.RuleName,
		RuleActive:          f.RuleActive,
		AgentPlatform:       f.AgentPlatform,
		CreatedBy:           f.CreatedBy,
		LastModifiedBy:      f.LastModifiedBy,
		CreatedDateGTE:      f.CreatedDateGTE,
		CreatedDateLTE:      f.CreatedDateLTE,
		LastModifiedDateGTE: f.LastModifiedDateGTE,
		LastModifiedDateLTE: f.LastModifiedDateLTE,
		SystemOwner:         f.SystemOwner,
		Params:              f.Params,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.RuleResponse, len(rules))
	for i := range rules {
		items[i] = *ruleToResponse(&rules[i])
	}
	return &database.List[dto.RuleResponse]{Items: items, Total: total}, nil
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
	return &dto.ResolveFilterValuesResponse{
		AgentPlatforms: platforms,
		Users:          users,
	}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func marshalConditions(filters []dto.FilterVM) (string, error) {
	b, err := json.Marshal(filters)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalConditions(raw string) []dto.FilterVM {
	if raw == "" {
		return nil
	}
	var filters []dto.FilterVM
	_ = json.Unmarshal([]byte(raw), &filters)
	return filters
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func templateDigestsToEntities(digests []dto.TemplateDigest) []domain.AlertResponseActionTemplate {
	templates := make([]domain.AlertResponseActionTemplate, 0, len(digests))
	for _, d := range digests {
		templates = append(templates, domain.AlertResponseActionTemplate{
			ID:          d.ID,
			Title:       d.Title,
			Description: d.Description,
			Command:     d.Command,
			SystemOwner: d.SystemOwner,
		})
	}
	return templates
}

func ruleToResponse(r *domain.AlertResponseRule) *dto.RuleResponse {
	resp := &dto.RuleResponse{
		ID:               r.ID,
		Name:             r.RuleName,
		Description:      r.RuleDescription,
		Command:          r.RuleCmd,
		Active:           r.RuleActive,
		AgentPlatform:    r.AgentPlatform,
		DefaultAgent:     r.DefaultAgent,
		Shell:            r.RuleShell,
		SystemOwner:      r.SystemOwner,
		CreatedBy:        r.CreatedBy,
		CreatedDate:      r.CreatedDate,
		LastModifiedBy:   r.LastModifiedBy,
		LastModifiedDate: r.LastModifiedDate,
		Conditions:       unmarshalConditions(r.RuleConditions),
	}
	if r.ExcludedAgents != "" {
		resp.ExcludedAgents = strings.Split(r.ExcludedAgents, ",")
	}
	for _, t := range r.Templates {
		resp.Actions = append(resp.Actions, dto.TemplateDigest{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Command:     t.Command,
			SystemOwner: t.SystemOwner,
		})
	}
	return resp
}
