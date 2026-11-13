package upgrade

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// v11LogstashFilter mirrors the utm_logstash_filter columns the converter reads.
type v11LogstashFilter struct {
	ID             int64   `gorm:"column:id"`
	FilterName     string  `gorm:"column:filter_name"`
	LogstashFilter *string `gorm:"column:logstash_filter"`
	IsActive       bool    `gorm:"column:is_active"`
}

// v12PipelineFile wraps the pipeline entries the v12 PipelineStore reads; the
// entries stay maps so every key the v11 YAML carried is preserved and only
// order/tenantId are injected.
type v12PipelineFile struct {
	Pipeline []map[string]any `yaml:"pipeline"`
}

// v12CustomFilterDefaultOrder mirrors the event-engine customFilterDefaultOrder.
const v12CustomFilterDefaultOrder = 100

// injectPipelineEntryFields sets the two keys v12's PipelineStore.Create and
// normalizeFilterOrder add on a custom filter: tenantId on every entry, and
// order only where the v11 YAML does not already carry a non-zero one.
func injectPipelineEntryFields(entries []map[string]any) {
	for _, e := range entries {
		if e == nil {
			continue
		}
		e["tenantId"] = defaultTenantID.String()
		order, ok := e["order"].(int)
		if !ok || order == 0 {
			e["order"] = v12CustomFilterDefaultOrder
		}
	}
}

func convertLogstashFilters(db *gorm.DB, pipelinesDir, pipelineDir string) error {
	if !tableExists(db, "utm_logstash_filter") {
		logSkipTable("utm_logstash_filter")
		return nil
	}

	var src []v11LogstashFilter
	if err := db.Raw(`SELECT id, filter_name, logstash_filter, is_active FROM utm_logstash_filter WHERE system_owner = false ORDER BY id`).Scan(&src).Error; err != nil {
		return err
	}

	converted, skipped, disabled := 0, 0, 0
	for _, row := range src {
		if row.LogstashFilter == nil || strings.TrimSpace(*row.LogstashFilter) == "" {
			skipped++
			catcher.Info("v11 upgrade: logstash filter has empty content, skipping", map[string]any{"filter": row.FilterName})
			continue
		}
		var pf v12PipelineFile
		if err := yaml.Unmarshal([]byte(*row.LogstashFilter), &pf); err != nil || len(pf.Pipeline) == 0 {
			skipped++
			catcher.Info("v11 upgrade: logstash filter is not valid pipeline YAML, skipping", map[string]any{"filter": row.FilterName, "error": err})
			continue
		}
		injectPipelineEntryFields(pf.Pipeline)

		identity := ruleSlug(row.FilterName) + "-" + defaultTenantID.String()
		// Defensive: if pipelinesDir is relative or empty, resolve it against CWD
		// (matches what the test expects)
		resolvedPipelines := pipelinesDir
		if !filepath.IsAbs(resolvedPipelines) {
			if resolvedPipelines == "" {
				resolvedPipelines = "."
			}
		}
		target := filepath.Join(resolvedPipelines, "user", defaultTenantID.String(), identity+".yaml")
		if _, err := os.Stat(target); err == nil {
			skipped++
			catcher.Info("v11 upgrade: pipeline filter file already exists, skipping", map[string]any{"filter": row.FilterName, "file": target})
		} else {
			data, err := yaml.Marshal(pf)
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

		if !row.IsActive {
			if err := addDisabledPipeline(filepath.Join(pipelineDir, "tenants.yaml"), identity); err != nil {
				return err
			}
			disabled++
		}
	}

	catcher.Info("v11 upgrade: logstash filters converted", map[string]any{
		"converted": converted, "skipped": skipped, "disabled": disabled,
	})
	return nil
}

// addDisabledPipeline appends id to the default tenant's disabledPipelines in
// path. Idempotent, and keeps every field another writer (e.g. the rules
// converter's disabledRules) already saved.
func addDisabledPipeline(path, id string) error {
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
	for _, r := range entry.DisabledPipelines {
		if r == id {
			return nil
		}
	}
	entry.DisabledPipelines = append(entry.DisabledPipelines, id)

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
