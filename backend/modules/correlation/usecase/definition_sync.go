package usecase

// TODO(deploy): ensure ./utmstack/rules directory is mounted in production

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gopkg.in/yaml.v3"
)

type DefinitionSyncService struct {
	ruleRepo     connectors.CorrelationRuleRepository
	dataTypeRepo connectors.DataTypeRepository
	rulesDir     string // default: "./utmstack/rules"; overridable via CORRELATION_RULES_DIR
}

func NewDefinitionSyncService(
	ruleRepo connectors.CorrelationRuleRepository,
	dataTypeRepo connectors.DataTypeRepository,
	rulesDir string,
) *DefinitionSyncService {
	return &DefinitionSyncService{
		ruleRepo:     ruleRepo,
		dataTypeRepo: dataTypeRepo,
		rulesDir:     rulesDir,
	}
}

type ruleYAML struct {
	ID            *int64   `yaml:"id"`
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	Adversary     string   `yaml:"adversary"`
	Category      string   `yaml:"category"`
	Technique     string   `yaml:"technique"`
	DataTypes     []string `yaml:"dataTypes"`
	References    []any    `yaml:"references"`
	Where         any      `yaml:"where"`
	AfterEvents   any      `yaml:"afterEvents"`
	GroupBy       []string `yaml:"groupBy"`
	DeduplicateBy []string `yaml:"deduplicateBy"`
	Impact        *struct {
		Confidentiality int `yaml:"confidentiality"`
		Integrity       int `yaml:"integrity"`
		Availability    int `yaml:"availability"`
	} `yaml:"impact"`
}

func (s *DefinitionSyncService) Sync(ctx context.Context) {
	logger.Info("DefinitionSyncService: starting rule sync from " + s.rulesDir)

	info, err := os.Stat(s.rulesDir)
	if err != nil || !info.IsDir() {
		logger.Warn("DefinitionSyncService: rules directory not found or not a directory: " + s.rulesDir + " — skipping sync")
		return
	}

	foundNames := make(map[string]bool)

	walkErr := filepath.Walk(s.rulesDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			logger.Error("DefinitionSyncService: error accessing path " + path + ": " + err.Error())
			return nil // continue walking
		}
		if fi.IsDir() {
			return nil // descend into subdirectories
		}
		if !isYAMLFile(fi.Name()) {
			return nil
		}

		rules, err := loadRulesFromFile(path)
		if err != nil {
			logger.Error("DefinitionSyncService: failed to load file " + path + ": " + err.Error())
			return nil
		}

		for _, ry := range rules {
			if ry.Name == "" {
				logger.Warn("DefinitionSyncService: skipping rule with empty name in file " + path)
				continue
			}

			foundNames[ry.Name] = true

			if err := s.upsertRule(ctx, ry); err != nil {
				logger.Error("DefinitionSyncService: failed to upsert rule " + ry.Name + ": " + err.Error())
			}
		}
		return nil
	})
	if walkErr != nil {
		logger.Error("DefinitionSyncService: directory walk failed: " + walkErr.Error())
		return
	}

	// Orphan cleanup — delete system-owned DB rules absent from disk.
	if err := s.deleteOrphanRules(ctx, foundNames); err != nil {
		logger.Error("DefinitionSyncService: orphan cleanup failed: " + err.Error())
	}

	logger.Info("DefinitionSyncService: rule sync completed")
}

func (s *DefinitionSyncService) upsertRule(ctx context.Context, ry ruleYAML) error {
	now := time.Now().UTC()

	// Resolve data types from DB by name.
	dataTypes := s.resolveDataTypes(ctx, ry.DataTypes)

	// Build the domain entity.
	rule := &domain.UtmCorrelationRules{
		RuleName:          ry.Name,
		RuleDescription:   ry.Description,
		RuleAdversary:     adversaryOrDefault(ry.Adversary),
		RuleCategory:      ry.Category,
		RuleTechnique:     ry.Technique,
		RuleDefinitionDef: marshalAny(ry.Where),
		RuleReferencesDef: marshalAny(ry.References),
		AfterEventsDef:    marshalAny(ry.AfterEvents),
		RuleGroupByDef:    marshalStrings(ry.GroupBy),
		DeduplicateByDef:  marshalStrings(ry.DeduplicateBy),
		RuleActive:        true,
		SystemOwner:       true,
		RuleLastUpdate:    &now,
		DataTypes:         dataTypes,
	}

	if ry.Impact != nil {
		rule.RuleConfidentiality = ry.Impact.Confidentiality
		rule.RuleIntegrity = ry.Impact.Integrity
		rule.RuleAvailability = ry.Impact.Availability
	}

	// Find existing by rule_name (system_owner=true).
	existing, err := s.ruleRepo.FindByRuleName(ctx, ry.Name)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update — preserve existing ID, keep system_owner=true.
		rule.ID = existing.ID
		return s.ruleRepo.Update(ctx, rule)
	}

	// Create new system-owned rule.
	return s.ruleRepo.Create(ctx, rule)
}

func (s *DefinitionSyncService) resolveDataTypes(ctx context.Context, names []string) []domain.UtmDataTypes {
	if len(names) == 0 {
		return nil
	}

	result := make([]domain.UtmDataTypes, 0, len(names))
	for _, name := range names {
		dt, err := s.dataTypeRepo.FindByDataType(ctx, name)
		if err != nil {
			logger.Warn("DefinitionSyncService: could not look up data type '" + name + "': " + err.Error())
			continue
		}
		if dt != nil {
			result = append(result, *dt)
		}
	}
	return result
}

func (s *DefinitionSyncService) deleteOrphanRules(ctx context.Context, foundNames map[string]bool) error {
	systemRules, err := s.ruleRepo.FindAllBySystemOwner(ctx, true)
	if err != nil {
		return err
	}

	for _, rule := range systemRules {
		if foundNames[rule.RuleName] {
			continue
		}
		logger.Info("DefinitionSyncService: deleting orphaned system rule: " + rule.RuleName)
		if err := s.ruleRepo.Delete(ctx, rule.ID); err != nil {
			logger.Error("DefinitionSyncService: failed to delete orphan rule " + rule.RuleName + ": " + err.Error())
		}
	}
	return nil
}

// ── file loaders ───────────────────────────────────────────────────────────────

func loadRulesFromFile(path string) ([]ruleYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try as list first.
	var list []ruleYAML
	if err := yaml.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return list, nil
	}

	// Try as single map.
	var single ruleYAML
	if err := yaml.Unmarshal(data, &single); err != nil {
		return nil, err
	}
	return []ruleYAML{single}, nil
}

func isYAMLFile(name string) bool {
	ext := filepath.Ext(name)
	return ext == ".yaml" || ext == ".yml"
}

// ── helpers ───────────────────────────────────────────────────────────────────

func adversaryOrDefault(adversary string) string {
	if adversary == "" {
		return string(domain.AdversaryOrigin)
	}
	return adversary
}

func marshalAny(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func marshalStrings(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	b, err := json.Marshal(ss)
	if err != nil {
		return ""
	}
	return string(b)
}
