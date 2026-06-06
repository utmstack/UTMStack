package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

// correlationRuleUsecase serves the rule CRUD/search from the file-backed
// RuleStore (YAML-direct). Identity is the rule's relative path.
type correlationRuleUsecase struct {
	store *RuleStore
}

func NewCorrelationRuleUsecase(store *RuleStore) connectors.CorrelationRuleUsecase {
	return &correlationRuleUsecase{store: store}
}

func (u *correlationRuleUsecase) Create(_ context.Context, req dto.CreateCorrelationRuleRequest) error {
	if len(req.DataTypes) == 0 {
		return domain.ErrDataTypesRequired
	}
	rule := buildRule(req.RuleName, req.RuleAdversary, req.RuleConfidentiality, req.RuleIntegrity,
		req.RuleAvailability, req.RuleCategory, req.RuleTechnique, req.RuleDescription,
		req.RuleReferencesDef, req.RuleDefinitionDef, req.AfterEventsDef, req.RuleGroupByDef,
		req.DeduplicateByDef, req.DataTypes)

	created, err := u.store.Create(rule)
	if err != nil {
		return mapStoreErr(err)
	}
	if !req.RuleActive {
		_ = u.store.SetEnabled(created.RelPath, false)
	}
	return nil
}

func (u *correlationRuleUsecase) Update(_ context.Context, req dto.UpdateCorrelationRuleRequest) error {
	if req.RelPath == "" {
		return domain.ErrIDRequired
	}
	if rawToWhere(req.RuleDefinitionDef) == "" {
		return domain.ErrCorrelationRuleNullDefinition
	}
	if len(req.DataTypes) == 0 {
		return domain.ErrDataTypesRequired
	}
	rule := buildRule(req.RuleName, req.RuleAdversary, req.RuleConfidentiality, req.RuleIntegrity,
		req.RuleAvailability, req.RuleCategory, req.RuleTechnique, req.RuleDescription,
		req.RuleReferencesDef, req.RuleDefinitionDef, req.AfterEventsDef, req.RuleGroupByDef,
		req.DeduplicateByDef, req.DataTypes)

	if _, err := u.store.Update(req.RelPath, rule); err != nil {
		return mapStoreErr(err)
	}
	// Reconcile the active state (the store keeps it in the filename).
	return mapStoreErr(u.store.SetEnabled(req.RelPath, req.RuleActive))
}

func (u *correlationRuleUsecase) GetByRelPath(_ context.Context, relPath string) (*dto.CorrelationRuleResponse, error) {
	sr := u.store.findByRelPath(relPath)
	if sr == nil {
		return nil, domain.ErrCorrelationRuleNotFound
	}
	return storedToResponse(sr), nil
}

func (u *correlationRuleUsecase) List(_ context.Context, f dto.CorrelationRuleFilters) (*connectors.ListResult[dto.CorrelationRuleResponse], error) {
	rules, total := u.store.List(RuleListFilter{
		Page:            f.Page,
		Size:            f.Size,
		Name:            f.RuleName,
		Search:          f.Search,
		Active:          f.RuleActive,
		SystemOwner:     f.SystemOwner,
		Categories:      f.RuleCategory,
		Adversaries:     f.RuleAdversary,
		Techniques:      f.RuleTechnique,
		Confidentiality: f.RuleConfidentiality,
		Integrity:       f.RuleIntegrity,
		Availability:    f.RuleAvailability,
		DataTypes:       f.DataTypes,
		InitDate:        parseISO(f.InitDate),
		EndDate:         parseISO(f.EndDate),
	})

	items := make([]dto.CorrelationRuleResponse, len(rules))
	for i, r := range rules {
		items[i] = *storedToResponse(r)
	}
	return &connectors.ListResult[dto.CorrelationRuleResponse]{Items: items, Total: int64(total)}, nil
}

func (u *correlationRuleUsecase) Delete(_ context.Context, relPath string) error {
	return mapStoreErr(u.store.Delete(relPath))
}

func (u *correlationRuleUsecase) SetActive(_ context.Context, relPath string, active bool) error {
	return mapStoreErr(u.store.SetEnabled(relPath, active))
}

func (u *correlationRuleUsecase) FindDistinctPropertyValues(_ context.Context, prop, value string) ([]string, error) {
	return u.store.DistinctValues(prop, value), nil
}

// ── mappers ───────────────────────────────────────────────────────────────────

func buildRule(name, adversary string, conf, integ, avail int, category, technique, description string,
	refs, def, after, groupBy, dedup json.RawMessage, dataTypes []dto.DataTypeRef) Rule {
	names := make([]string, 0, len(dataTypes))
	for _, d := range dataTypes {
		names = append(names, d.DataType)
	}
	return Rule{
		Name:          name,
		Adversary:     adversary,
		Category:      category,
		Technique:     technique,
		Description:   description,
		DataTypes:     names,
		Impact:        Impact{Confidentiality: conf, Integrity: integ, Availability: avail},
		Where:         rawToWhere(def),
		References:    rawToAnySlice(refs),
		AfterEvents:   rawToAny(after),
		GroupBy:       rawToStrSlice(groupBy),
		DeduplicateBy: rawToStrSlice(dedup),
	}
}

func storedToResponse(sr *StoredRule) *dto.CorrelationRuleResponse {
	mod := sr.Modified
	dataTypes := make([]dto.RuleDataTypeResponse, 0, len(sr.DataTypes))
	for _, name := range sr.DataTypes {
		dataTypes = append(dataTypes, dto.RuleDataTypeResponse{DataType: name, DataTypeName: name})
	}
	return &dto.CorrelationRuleResponse{
		RelPath:             sr.RelPath,
		RuleName:            sr.Name,
		RuleAdversary:       sr.Adversary,
		RuleConfidentiality: sr.Impact.Confidentiality,
		RuleIntegrity:       sr.Impact.Integrity,
		RuleAvailability:    sr.Impact.Availability,
		RuleCategory:        sr.Category,
		RuleTechnique:       sr.Technique,
		RuleDescription:     sr.Description,
		RuleReferencesDef:   anyToRaw(sr.References),
		RuleDefinitionDef:   anyToRaw(sr.Where),
		AfterEventsDef:      anyToRaw(sr.AfterEvents),
		RuleGroupByDef:      anyToRaw(sr.GroupBy),
		DeduplicateByDef:    anyToRaw(sr.DeduplicateBy),
		RuleLastUpdate:      &mod,
		RuleActive:          sr.Active(),
		SystemOwner:         sr.SystemOwned(),
		DataTypes:           dataTypes,
	}
}

func mapStoreErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrRuleNotFound):
		return domain.ErrCorrelationRuleNotFound
	case errors.Is(err, ErrSystemRuleContent):
		return domain.ErrCorrelationRuleSystemOwner
	default:
		return err
	}
}

// rawToWhere extracts the CEL string from the request's `definition`. The field
// is normally a JSON string; if it is not, the raw bytes are used verbatim.
func rawToWhere(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

func rawToAnySlice(raw json.RawMessage) []any {
	if len(raw) == 0 {
		return nil
	}
	var out []any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func rawToStrSlice(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func rawToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func anyToRaw(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func parseISO(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
