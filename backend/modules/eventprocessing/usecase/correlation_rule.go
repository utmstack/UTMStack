package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"gopkg.in/yaml.v3"
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

	correlate,err:= req.GetCorrelationDef()
	if err!=nil{
		return err
	}

	if err := validateRuleContent(req.RuleDefinitionDef,correlate ); err != nil {
		return err
	}
	rule := buildRule(req.RuleName, req.RuleAdversary, req.RuleConfidentiality, req.RuleIntegrity,
		req.RuleAvailability, req.RuleCategory, req.RuleTechnique, req.RuleDescription,
		req.RuleReferencesDef, req.RuleDefinitionDef, correlate, req.RuleGroupByDef,
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

func (u *correlationRuleUsecase) ImportRules(_ context.Context, files []dto.ImportRuleFile) []dto.ImportRuleResult {
	results := make([]dto.ImportRuleResult, 0, len(files))
	for _, f := range files {
		res := dto.ImportRuleResult{Filename: f.Filename}

		rule, err := parseRuleYAML([]byte(f.Content))
		if err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}
		res.Name = rule.Name
		if err := validateImportedRule(rule); err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}
		created, err := u.store.Create(rule)
		if err != nil {
			res.Error = mapStoreErr(err).Error()
			results = append(results, res)
			continue
		}
		res.Approved = true
		res.RelPath = created.RelPath
		results = append(results, res)
	}
	return results
}

// parseRuleYAML accepts both the canonical list form (`- dataTypes: ...`, how
// rule files are stored) and a single mapping.
func parseRuleYAML(data []byte) (Rule, error) {
	if strings.TrimSpace(string(data)) == "" {
		return Rule{}, fmt.Errorf("%w: file is empty", domain.ErrCorrelationRuleInvalidContent)
	}
	var list []Rule
	if err := yaml.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return list[0], nil
	}
	var single Rule
	if err := yaml.Unmarshal(data, &single); err != nil {
		return Rule{}, fmt.Errorf("%w: not valid rule YAML: %v", domain.ErrCorrelationRuleInvalidContent, err)
	}
	return single, nil
}

// validateImportedRule checks a parsed rule has the fields the Event Processor
// needs, reusing the same afterEvents step-shape validation as create/update.
func validateImportedRule(r Rule) error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("%w: missing rule name", domain.ErrCorrelationRuleInvalidContent)
	}
	if len(r.DataTypes) == 0 {
		return domain.ErrDataTypesRequired
	}
	if strings.TrimSpace(r.Where) == "" {
		return domain.ErrCorrelationRuleNullDefinition
	}
	if r.Correlation == nil {
		return nil
	}
	raw, err := json.Marshal(r.Correlation)
	if err != nil {
		return fmt.Errorf("%w: afterEvents is not serializable", domain.ErrCorrelationRuleInvalidContent)
	}
	var steps []validateStep
	if err := json.Unmarshal(raw, &steps); err != nil {
		return fmt.Errorf("%w: afterEvents must be an array of correlation steps", domain.ErrCorrelationRuleInvalidContent)
	}
	for i := range steps {
		if err := validateStepShape(&steps[i], i+1); err != nil {
			return err
		}
	}
	return nil
}

func (u *correlationRuleUsecase) Update(_ context.Context, req dto.UpdateCorrelationRuleRequest) error {
	if req.RelPath == "" {
		return domain.ErrIDRequired
	}
	if len(req.DataTypes) == 0 {
		return domain.ErrDataTypesRequired
	}

	correlate,err:= req.GetCorrelationDef()
	if err!=nil{
		return err
	}

	if err := validateRuleContent(req.RuleDefinitionDef, correlate); err != nil {
		return err
	}
	rule := buildRule(req.RuleName, req.RuleAdversary, req.RuleConfidentiality, req.RuleIntegrity,
		req.RuleAvailability, req.RuleCategory, req.RuleTechnique, req.RuleDescription,
		req.RuleReferencesDef, req.RuleDefinitionDef, correlate, req.RuleGroupByDef,
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
		Correlation:   rawToAny(after),
		GroupBy:       rawToStrSlice(groupBy),
		DeduplicateBy: rawToStrSlice(dedup),
	}
}

func storedToResponse(sr *StoredRule) *dto.CorrelationRuleResponse {
	mod := sr.Modified
	dataTypes := make([]dto.RuleDataTypeResponse, 0, len(sr.DataTypes))
	for _, name := range sr.DataTypes {
		// A rule's stored dataTypes are exactly the ones it targets, so they are
		// included. Without this the UI (which filters by `included`) drops them
		// and a subsequent save would persist an empty dataTypes list.
		dataTypes = append(dataTypes, dto.RuleDataTypeResponse{DataType: name, DataTypeName: name, Included: true})
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
		CorrelationDef:      anyToRaw(sr.Correlation),
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

var validAfterEventOps = map[string]bool{
	"filter_term": true, "must_not_term": true, "filter_match": true, "must_not_match": true,
}

type validateCond struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
}
type validateStep struct {
	IndexPattern string         `json:"indexPattern"`
	Within       string         `json:"within"`
	With         []validateCond `json:"with"`
	Or           []validateStep `json:"or"`
}

func validateRuleContent(def, after json.RawMessage) error {
	if rawToWhere(def) == "" {
		return domain.ErrCorrelationRuleNullDefinition
	}
	if len(after) == 0 || string(after) == "null" {
		return nil
	}
	var steps []validateStep
	if err := json.Unmarshal(after, &steps); err != nil {
		return fmt.Errorf("%w: afterEvents must be an array of correlation steps", domain.ErrCorrelationRuleInvalidContent)
	}
	for i := range steps {
		if err := validateStepShape(&steps[i], i+1); err != nil {
			return err
		}
	}
	return nil
}

func validateStepShape(s *validateStep, n int) error {
	if strings.TrimSpace(s.IndexPattern) == "" {
		return fmt.Errorf("%w: step %d needs an index pattern", domain.ErrCorrelationRuleInvalidContent, n)
	}
	if s.Within != "" {
		if _, err := time.ParseDuration(s.Within); err != nil {
			return fmt.Errorf("%w: step %d 'within' (%q) is not a valid duration", domain.ErrCorrelationRuleInvalidContent, n, s.Within)
		}
	}
	for _, c := range s.With {
		if strings.TrimSpace(c.Field) == "" {
			return fmt.Errorf("%w: step %d has a condition without a field", domain.ErrCorrelationRuleInvalidContent, n)
		}
		if !validAfterEventOps[c.Operator] {
			return fmt.Errorf("%w: step %d uses an unknown operator %q", domain.ErrCorrelationRuleInvalidContent, n, c.Operator)
		}
	}
	for i := range s.Or {
		if err := validateStepShape(&s.Or[i], n); err != nil {
			return err
		}
	}
	return nil
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
