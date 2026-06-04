package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type alertTagRuleUsecase struct {
	ruleRepo connectors.AlertTagRuleRepository
	tagRepo  connectors.AlertTagRepository
}

func NewAlertTagRuleUsecase(
	ruleRepo connectors.AlertTagRuleRepository,
	tagRepo connectors.AlertTagRepository,
) connectors.AlertTagRuleUsecase {
	return &alertTagRuleUsecase{ruleRepo: ruleRepo, tagRepo: tagRepo}
}

func (u *alertTagRuleUsecase) Create(ctx context.Context, req dto.CreateAlertTagRuleRequest) (*dto.AlertTagRuleResponse, error) {
	condJSON, err := json.Marshal(req.Conditions)
	if err != nil {
		return nil, fmt.Errorf("marshal conditions: %w", err)
	}

	appliedTags := tagsToCSV(req.Tags)

	rule := domain.UtmAlertTagRule{
		RuleName:        req.Name,
		RuleDescription: req.Description,
		RuleConditions:  string(condJSON),
		RuleAppliedTags: appliedTags,
		RuleActive:      true,
		RuleDeleted:     false,
	}

	created, err := u.ruleRepo.Create(ctx, rule)
	if err != nil {
		return nil, err
	}

	return u.toResponse(ctx, created)
}

func (u *alertTagRuleUsecase) Update(ctx context.Context, req dto.UpdateAlertTagRuleRequest) (*dto.AlertTagRuleResponse, error) {
	existing, err := u.ruleRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrAlertTagRuleNotFound
	}

	condJSON, err := json.Marshal(req.Conditions)
	if err != nil {
		return nil, fmt.Errorf("marshal conditions: %w", err)
	}

	existing.RuleName = req.Name
	existing.RuleDescription = req.Description
	existing.RuleConditions = string(condJSON)
	existing.RuleAppliedTags = tagsToCSV(req.Tags)

	updated, err := u.ruleRepo.Update(ctx, *existing)
	if err != nil {
		return nil, err
	}

	return u.toResponse(ctx, updated)
}

func (u *alertTagRuleUsecase) List(ctx context.Context, filters dto.AlertTagRuleFilters) ([]dto.AlertTagRuleResponse, int64, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Size < 1 || filters.Size > 200 {
		filters.Size = 20
	}

	rows, total, err := u.ruleRepo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.AlertTagRuleResponse, 0, len(rows))
	for i := range rows {
		resp, err := u.toResponse(ctx, &rows[i])
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *resp)
	}

	return result, total, nil
}

func (u *alertTagRuleUsecase) GetByID(ctx context.Context, id uint64) (*dto.AlertTagRuleResponse, error) {
	row, err := u.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, domain.ErrAlertTagRuleNotFound
	}
	return u.toResponse(ctx, row)
}

func (u *alertTagRuleUsecase) GetByIDs(ctx context.Context, ids []int64) ([]dto.AlertTagRuleResponse, error) {
	rows, err := u.ruleRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make([]dto.AlertTagRuleResponse, 0, len(rows))
	for i := range rows {
		resp, err := u.toResponse(ctx, &rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *resp)
	}
	return result, nil
}

func (u *alertTagRuleUsecase) Delete(ctx context.Context, id uint64) error {
	existing, err := u.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrAlertTagRuleNotFound
	}
	return u.ruleRepo.Delete(ctx, id)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (u *alertTagRuleUsecase) toResponse(ctx context.Context, rule *domain.UtmAlertTagRule) (*dto.AlertTagRuleResponse, error) {
	var conditions []common_models.FilterType
	if rule.RuleConditions != "" {
		if err := json.Unmarshal([]byte(rule.RuleConditions), &conditions); err != nil {
			conditions = nil
		}
	}

	tagRefs, err := u.resolveTagRefs(ctx, rule.RuleAppliedTags)
	if err != nil {
		return nil, err
	}

	return &dto.AlertTagRuleResponse{
		ID:               rule.ID,
		Name:             rule.RuleName,
		Description:      rule.RuleDescription,
		Conditions:       conditions,
		Tags:             tagRefs,
		Active:           rule.RuleActive,
		Deleted:          rule.RuleDeleted,
		CreatedBy:        rule.CreatedBy,
		CreatedDate:      rule.CreatedDate,
		LastModifiedBy:   rule.LastModifiedBy,
		LastModifiedDate: rule.LastModifiedDate,
	}, nil
}

func (u *alertTagRuleUsecase) resolveTagRefs(ctx context.Context, csv string) ([]dto.AlertTagRuleTagRef, error) {
	ids := csvToInt64Slice(csv)
	if len(ids) == 0 {
		return nil, nil
	}

	tags, err := u.tagRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	refs := make([]dto.AlertTagRuleTagRef, 0, len(tags))
	for _, t := range tags {
		refs = append(refs, dto.AlertTagRuleTagRef{
			ID:          t.ID,
			TagName:     t.TagName,
			TagColor:    t.TagColor,
			SystemOwner: t.SystemOwner,
		})
	}
	return refs, nil
}

func tagsToCSV(tags []dto.AlertTagRuleTagRef) string {
	parts := make([]string, 0, len(tags))
	for _, t := range tags {
		parts = append(parts, fmt.Sprintf("%d", t.ID))
	}
	return strings.Join(parts, ",")
}

func csvToInt64Slice(csv string) []int64 {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var n int64
		if _, err := fmt.Sscanf(p, "%d", &n); err == nil {
			ids = append(ids, n)
		}
	}
	return ids
}
