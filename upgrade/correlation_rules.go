package upgrade

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// v11CorrelationRule mirrors the utm_correlation_rules columns the converter
// reads. The *_def columns are raw JSON strings; NULL means "empty".
// DataTypes comes from a subquery over the join tables when they exist.
type v11CorrelationRule struct {
	ID                   int64
	RuleName             string
	RuleConfidentiality  int
	RuleIntegrity        int
	RuleAvailability     int
	RuleCategory         string
	RuleTechnique        string
	RuleAdversary        string
	RuleDescription      *string
	RuleReferencesDef    *string
	RuleDefinitionDef    string
	RuleActive           bool
	RuleAfterEventsDef   *string
	RuleGroupByDef       *string
	RuleDeduplicateByDef *string
	DataTypes            string `gorm:"column:data_types"`
}

// v12Rule mirrors the event-engine domain.Rule yaml tags so the converter
// writes exactly what RuleStore reads.
type v12Rule struct {
	DataTypes     []string  `yaml:"dataTypes"`
	Name          string    `yaml:"name"`
	Impact        v12Impact `yaml:"impact"`
	Category      string    `yaml:"category"`
	Technique     string    `yaml:"technique"`
	Adversary     string    `yaml:"adversary"`
	References    []any     `yaml:"references,omitempty"`
	Description   string    `yaml:"description,omitempty"`
	Where         string    `yaml:"where"`
	Correlation   any       `yaml:"correlation,omitempty"`
	GroupBy       []string  `yaml:"groupBy,omitempty"`
	DeduplicateBy []string  `yaml:"deduplicateBy,omitempty"`
	TenantId      string    `yaml:"tenantId,omitempty"`
}

type v12Impact struct {
	Confidentiality int `yaml:"confidentiality"`
	Integrity       int `yaml:"integrity"`
	Availability    int `yaml:"availability"`
}

// tenantsFile mirrors the event-engine domain.TenantsFile. Every field is kept
// so a read-modify-write never drops entries another writer added.
type tenantsFile struct {
	Tenants []tenantEntry `yaml:"tenants"`
}

type tenantEntry struct {
	ID                string   `yaml:"id"`
	Name              string   `yaml:"name"`
	Assets            []any    `yaml:"assets,omitempty"`
	DisabledRules     []string `yaml:"disabledRules,omitempty"`
	DisabledPipelines []string `yaml:"disabledPipelines,omitempty"`
	PipelineOrder     []string `yaml:"pipelineOrder,omitempty"`
}

// defaultTenantYAMLName must match the event-engine's DefaultTenantName.
const defaultTenantYAMLName = "Default"

var (
	ruleSlugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	// v11 field prefixes utm.* / event.* are log.* in v12; the rest of the
	// CEL string is untouched.
	fieldPrefixRe = regexp.MustCompile(`\b(utm|event)\.`)
)

// ruleSlug reproduces the event-engine RuleStore slug algorithm.
func ruleSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = ruleSlugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "rule"
	}
	return s
}

func jsonAnySlice(s *string) []any {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	var out []any
	if err := json.Unmarshal([]byte(*s), &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func jsonStrSlice(s *string) []string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(*s), &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// rewriteCorrelationFields replaces utm./event. prefixes in every `field`
// value of the correlation structure, including nested with[] and or[] lists.
func rewriteCorrelationFields(entries []any) {
	for _, entry := range entries {
		m, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if f, ok := m["field"].(string); ok {
			m["field"] = fieldPrefixRe.ReplaceAllString(f, "log.")
		}
		if with, ok := m["with"].([]any); ok {
			rewriteCorrelationFields(with)
		}
		if or, ok := m["or"].([]any); ok {
			rewriteCorrelationFields(or)
		}
	}
}

// addDisabledRule appends id to the default tenant's disabledRules in path,
// creating the file when it is missing. Idempotent.
func addDisabledRule(path, id string) error {
	tf := tenantsFile{}
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &tf); err != nil {
			return err
		}
	}

	var entry *tenantEntry
	for i := range tf.Tenants {
		if tf.Tenants[i].ID == defaultTenantID.String() {
			entry = &tf.Tenants[i]
			break
		}
	}
	if entry == nil {
		tf.Tenants = append(tf.Tenants, tenantEntry{ID: defaultTenantID.String(), Name: defaultTenantYAMLName})
		entry = &tf.Tenants[len(tf.Tenants)-1]
	}
	for _, r := range entry.DisabledRules {
		if r == id {
			return nil
		}
	}
	entry.DisabledRules = append(entry.DisabledRules, id)

	data, err := yaml.Marshal(tf)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func convertCorrelationRules(db *gorm.DB, rulesDir, pipelineDir string) error {
	if !tableExists(db, "utm_correlation_rules") {
		logSkipTable("utm_correlation_rules")
		return nil
	}

	query := `SELECT r.id, r.rule_name, r.rule_confidentiality, r.rule_integrity, r.rule_availability,
		r.rule_category, r.rule_technique, r.rule_adversary, r.rule_description, r.rule_references_def,
		r.rule_definition_def, r.rule_active, r.rule_after_events_def, r.rule_group_by_def, r.rule_deduplicate_by_def`
	if tableExists(db, "utm_group_rules_data_type") && tableExists(db, "utm_data_types") {
		query += `, array_to_string((SELECT array_agg(DISTINCT dt.data_type) FROM utm_group_rules_data_type gr
			JOIN utm_data_types dt ON dt.id = gr.data_type_id WHERE gr.rule_id = r.id), ',') AS data_types`
	}
	query += ` FROM utm_correlation_rules r WHERE r.system_owner = false ORDER BY r.id`

	var src []v11CorrelationRule
	if err := db.Raw(query).Scan(&src).Error; err != nil {
		return err
	}

	converted, skipped, disabled := 0, 0, 0
	for _, row := range src {
		rule := v12Rule{
			DataTypes:   splitDataTypes(row.DataTypes),
			Name:        row.RuleName,
			Impact:      v12Impact{Confidentiality: row.RuleConfidentiality, Integrity: row.RuleIntegrity, Availability: row.RuleAvailability},
			Category:    row.RuleCategory,
			Technique:   row.RuleTechnique,
			Adversary:   row.RuleAdversary,
			References:  jsonAnySlice(row.RuleReferencesDef),
			Description: strFrom(row.RuleDescription),
			Where:       fieldPrefixRe.ReplaceAllString(row.RuleDefinitionDef, "log."),
			TenantId:    defaultTenantID.String(),
		}
		if row.RuleAfterEventsDef != nil && strings.TrimSpace(*row.RuleAfterEventsDef) != "" {
			if corr := jsonAnySlice(row.RuleAfterEventsDef); corr != nil {
				rewriteCorrelationFields(corr)
				rule.Correlation = corr
			}
		}
		if row.RuleGroupByDef != nil {
			rule.GroupBy = rewriteStrings(jsonStrSlice(row.RuleGroupByDef))
		}
		if row.RuleDeduplicateByDef != nil {
			rule.DeduplicateBy = rewriteStrings(jsonStrSlice(row.RuleDeduplicateByDef))
		}

		identity := ruleSlug(row.RuleName) + "-" + defaultTenantID.String()
		target := filepath.Join(rulesDir, "user", defaultTenantID.String(), identity+".yaml")
		if _, err := os.Stat(target); err == nil {
			skipped++
			catcher.Info("v11 upgrade: correlation rule file already exists, skipping", map[string]any{"rule": row.RuleName, "file": target})
		} else {
			data, err := yaml.Marshal(rule)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(target, data, 0o644); err != nil {
				return err
			}
			converted++
		}

		if !row.RuleActive {
			if err := addDisabledRule(filepath.Join(pipelineDir, "tenants.yaml"), identity); err != nil {
				return err
			}
			disabled++
		}
	}

	catcher.Info("v11 upgrade: correlation rules converted", map[string]any{
		"converted": converted, "skipped": skipped, "disabled": disabled,
	})
	return nil
}

func rewriteStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	for i, s := range in {
		in[i] = fieldPrefixRe.ReplaceAllString(s, "log.")
	}
	return in
}

func splitDataTypes(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) == 0 {
		return nil
	}
	return parts
}
