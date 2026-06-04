package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	"github.com/utmstack/utmstack/backend/modules/correlation/dto"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
)

type correlationRuleUsecase struct {
	repo connectors.CorrelationRuleRepository
}

func NewCorrelationRuleUsecase(repo connectors.CorrelationRuleRepository) connectors.CorrelationRuleUsecase {
	return &correlationRuleUsecase{repo: repo}
}

func (u *correlationRuleUsecase) Create(ctx context.Context, req dto.CreateCorrelationRuleRequest) error {
	if len(req.DataTypes) == 0 {
		return correrrors.ErrDataTypesRequired
	}

	now := time.Now().UTC()

	rule := &domain.UtmCorrelationRules{
		RuleName:            req.RuleName,
		RuleAdversary:       req.RuleAdversary,
		RuleConfidentiality: req.RuleConfidentiality,
		RuleIntegrity:       req.RuleIntegrity,
		RuleAvailability:    req.RuleAvailability,
		RuleCategory:        req.RuleCategory,
		RuleTechnique:       req.RuleTechnique,
		RuleDescription:     req.RuleDescription,
		RuleReferencesDef:   rawToString(req.RuleReferencesDef),
		RuleDefinitionDef:   rawToString(req.RuleDefinitionDef),
		AfterEventsDef:      rawToString(req.AfterEventsDef),
		RuleGroupByDef:      rawToString(req.RuleGroupByDef),
		DeduplicateByDef:    rawToString(req.DeduplicateByDef),
		RuleActive:          req.RuleActive,
		SystemOwner:         false, // always false in prod — no isInDevelop
		RuleLastUpdate:      &now,
		DataTypes:           dataTypeRefsToEntities(req.DataTypes),
	}

	return u.repo.Create(ctx, rule)
}

func (u *correlationRuleUsecase) Update(ctx context.Context, req dto.UpdateCorrelationRuleRequest) error {
	if req.ID == nil || *req.ID == 0 {
		return correrrors.ErrIDRequired
	}

	// Guard 1: definition must not be empty.
	if rawToString(req.RuleDefinitionDef) == "" {
		return correrrors.ErrCorrelationRuleNullDefinition
	}

	// Guard 2: rule must exist (checked before dataTypes to mirror Java's EntityNotFoundException → 404).
	existing, err := u.repo.GetByID(ctx, *req.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return correrrors.ErrCorrelationRuleNotFound
	}

	// Guard 3: must have at least one data type.
	if len(req.DataTypes) == 0 {
		return correrrors.ErrDataTypesRequired
	}

	// Guard 4: system rules cannot be updated.
	if existing.SystemOwner {
		return correrrors.ErrCorrelationRuleSystemOwner
	}

	now := time.Now().UTC()
	rule := &domain.UtmCorrelationRules{
		ID:                  *req.ID,
		RuleName:            req.RuleName,
		RuleAdversary:       req.RuleAdversary,
		RuleConfidentiality: req.RuleConfidentiality,
		RuleIntegrity:       req.RuleIntegrity,
		RuleAvailability:    req.RuleAvailability,
		RuleCategory:        req.RuleCategory,
		RuleTechnique:       req.RuleTechnique,
		RuleDescription:     req.RuleDescription,
		RuleReferencesDef:   rawToString(req.RuleReferencesDef),
		RuleDefinitionDef:   rawToString(req.RuleDefinitionDef),
		AfterEventsDef:      rawToString(req.AfterEventsDef),
		RuleGroupByDef:      rawToString(req.RuleGroupByDef),
		DeduplicateByDef:    rawToString(req.DeduplicateByDef),
		RuleActive:          req.RuleActive,
		SystemOwner:         existing.SystemOwner,
		RuleLastUpdate:      &now,
		DataTypes:           dataTypeRefsToEntities(req.DataTypes),
	}

	return u.repo.Update(ctx, rule)
}

func (u *correlationRuleUsecase) GetByID(ctx context.Context, id int64) (*dto.CorrelationRuleResponse, error) {
	rule, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, correrrors.ErrCorrelationRuleNotFound
	}
	return dto.CorrelationRuleToResponse(rule), nil
}

func (u *correlationRuleUsecase) List(ctx context.Context, f dto.CorrelationRuleFilters) (*connectors.ListResult[dto.CorrelationRuleResponse], error) {
	page, size := normPage(f.Page, f.Size)

	rules, total, err := u.repo.List(ctx, connectors.CorrelationRuleFilters{
		Page:                page,
		Size:                size,
		RuleName:            f.RuleName,
		RuleActive:          f.RuleActive,
		RuleCategory:        f.RuleCategory,
		RuleAdversary:       f.RuleAdversary,
		RuleTechnique:       f.RuleTechnique,
		RuleConfidentiality: f.RuleConfidentiality,
		RuleIntegrity:       f.RuleIntegrity,
		RuleAvailability:    f.RuleAvailability,
		SystemOwner:         f.SystemOwner,
		DataTypes:           f.DataTypes,
		InitDate:            f.InitDate,
		EndDate:             f.EndDate,
		Search:              f.Search,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.CorrelationRuleResponse, len(rules))
	for i := range rules {
		items[i] = *dto.CorrelationRuleToResponse(&rules[i])
	}
	return &connectors.ListResult[dto.CorrelationRuleResponse]{Items: items, Total: total}, nil
}

func (u *correlationRuleUsecase) Delete(ctx context.Context, id int64) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return correrrors.ErrCorrelationRuleNotFound
	}
	if existing.SystemOwner {
		return correrrors.ErrCorrelationRuleSystemOwner
	}
	return u.repo.Delete(ctx, id)
}

func (u *correlationRuleUsecase) ActivateDeactivate(ctx context.Context, id int64, active bool) error {
	return u.repo.ActivateDeactivate(ctx, id, active)
}

func (u *correlationRuleUsecase) FindDistinctPropertyValues(ctx context.Context, prop, value string) ([]string, error) {
	return u.repo.FindDistinctPropertyValues(ctx, prop, value)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func rawToString(r []byte) string {
	if len(r) == 0 {
		return ""
	}
	return string(r)
}

func dataTypeRefsToEntities(refs []dto.DataTypeRef) []domain.UtmDataTypes {
	result := make([]domain.UtmDataTypes, 0, len(refs))
	for _, r := range refs {
		dt := domain.UtmDataTypes{
			DataType:            r.DataType,
			DataTypeName:        r.DataTypeName,
			DataTypeDescription: r.DataTypeDescription,
			Included:            r.Included,
		}
		if r.ID != nil {
			dt.ID = *r.ID
		}
		result = append(result, dt)
	}
	return result
}
