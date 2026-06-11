package usecase

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"gorm.io/gorm"
)

// PipelineBootstrap migrates utm_tenant_config and utm_regex_pattern from the
// legacy DB into their respective pipeline YAML files, then drops the tables.
type PipelineBootstrap struct {
	writer *pipelineWriter
	db     *gorm.DB
}

func NewPipelineBootstrap(writer *pipelineWriter, db *gorm.DB) *PipelineBootstrap {
	return &PipelineBootstrap{writer: writer, db: db}
}

// Run is idempotent and safe to call on every boot.
func (b *PipelineBootstrap) Run(ctx context.Context) error {
	if err := b.migrateTenantConfig(ctx); err != nil {
		_ = catcher.Error("eventprocessing: tenant config migration failed", err, nil)
	}
	if err := b.migrateRegexPatterns(ctx); err != nil {
		_ = catcher.Error("eventprocessing: regex pattern migration failed", err, nil)
	}
	return nil
}

func (b *PipelineBootstrap) migrateTenantConfig(ctx context.Context) error {
	if !b.db.Migrator().HasTable(&domain.UtmTenantConfig{}) {
		// Fresh install — write empty default tenant so the engine always has a valid file.
		return b.writer.WriteTenants(nil)
	}

	var rows []domain.UtmTenantConfig
	if err := b.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}

	assets := make([]assetFileYAML, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		a := assetFileYAML{
			Name:            r.AssetName,
			Confidentiality: uint32(r.AssetConfidentiality),
			Availability:    uint32(r.AssetAvailability),
			Integrity:       uint32(r.AssetIntegrity),
		}
		a.Hostnames = deserializeStringList(r.AssetHostnameListDef)
		a.Ips = deserializeStringList(r.AssetIpListDef)
		assets = append(assets, a)
	}

	if err := b.writer.WriteTenants(assets); err != nil {
		return err
	}

	if b.db.Migrator().HasTable("datasources") {
		for i := range rows {
			r := &rows[i]
			names := append(deserializeStringList(r.AssetHostnameListDef), r.AssetName)
			ips := deserializeStringList(r.AssetIpListDef)
			match := b.db.Where("asset_name IN ?", names)
			if len(ips) > 0 {
				match = match.Or("asset_ip IN ?", ips)
			}
			if err := b.db.WithContext(ctx).Table("datasources").Where(match).
				Updates(map[string]any{
					"asset_confidentiality": r.AssetConfidentiality,
					"asset_integrity":       r.AssetIntegrity,
					"asset_availability":    r.AssetAvailability,
				}).Error; err != nil {
				_ = catcher.Error("eventprocessing: carrying legacy asset CIA to datasource failed", err, map[string]any{"asset": r.AssetName})
			}
		}
	}

	if err := b.db.Exec("DROP TABLE IF EXISTS utm_tenant_config CASCADE").Error; err != nil {
		return err
	}
	catcher.Info("eventprocessing: utm_tenant_config migrated to datasource CIA + tenants.yaml and dropped", nil)
	return nil
}

func (b *PipelineBootstrap) migrateRegexPatterns(ctx context.Context) error {
	// Start with the system patterns (baked in; always present).
	patterns := make(map[string]string, len(systemPatterns))
	for k, v := range systemPatterns {
		patterns[k] = v
	}

	if !b.db.Migrator().HasTable(&domain.UtmRegexPattern{}) {
		// Fresh install — write system patterns only.
		return b.writer.WritePatterns(patterns)
	}

	// Upgrade: merge user-created patterns from the DB (system rows already in
	// patterns map above; user rows override only if their patternId differs).
	var rows []domain.UtmRegexPattern
	if err := b.db.WithContext(ctx).Where("system_owner = false").Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		patterns[rows[i].PatternID] = rows[i].PatternDefinition
	}

	if err := b.writer.WritePatterns(patterns); err != nil {
		return err
	}

	if err := b.db.Exec("DROP TABLE IF EXISTS utm_regex_pattern CASCADE").Error; err != nil {
		return err
	}
	catcher.Info("eventprocessing: utm_regex_pattern migrated to patterns.yaml and dropped", nil)
	return nil
}

// systemPatterns are the 26 canonical named patterns the filter engine uses as
// {{.patternName}} references. Previously seeded into utm_regex_pattern; now
// baked into the bootstrap so the patterns.yaml always contains them even on a
// fresh install (no DB table needed).
var systemPatterns = map[string]string{
	"ciscoMacAddr":    `(?:(?:[A-Fa-f0-9]{4}\.){2}[A-Fa-f0-9]{4})`,
	"syslogDate":      `[A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2}`,
	"winMacAddr":      `(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})`,
	"commonMacAddr":   `(?:(?:[A-Fa-f0-9]{2}:){5}[A-Fa-f0-9]{2})|(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})`,
	"integer":         `(?:[+-]?(?:[0-9]+))`,
	"day":             `(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)`,
	"word":            `\b\w+\b`,
	"greedy":          `.*`,
	"space":           `\s+`,
	"notSpace":        `\S+`,
	"monthName":       `\b(?:[Jj]an(?:uary|uar)?|[Ff]eb(?:ruary|ruar)?|[Mm](?:a|ä)?r(?:ch|z)?|[Aa]pr(?:il)?|[Mm]a(?:y|i)?|[Jj]un(?:e|i)?|[Jj]ul(?:y|i)?|[Aa]ug(?:ust)?|[Ss]ep(?:tember)?|[Oo](?:c|k)?t(?:ober)?|[Nn]ov(?:ember)?|[Dd]e(?:c|z)(?:ember)?)\b`,
	"ipv4":            `(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.)){3}((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))`,
	"email":           `((?P<name>[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+)@(?P<domain>[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*))`,
	"domain":          `((?:[_a-z0-9](?:[_a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?)?)`,
	"hostname":        `(\b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(\.?|\b))`,
	"data":            `(.*?)`,
	"ipv6":            `([0-9a-fA-F]{1,4}(:[0-9a-fA-F]{0,4}){1,7}|::[0-1]?)`,
	"uuid":            `([A-Fa-f0-9]{8}-(?:[A-Fa-f0-9]{4}-){3}[A-Fa-f0-9]{12})`,
	"monthNumber":     `(?:0[1-9]|1[0-2])`,
	"monthDay":        `(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])`,
	"year":            `(([1-9])[0-9]{1,3})`,
	"hour":            `(([01][0-9])|2[0-4])`,
	"minute":          `(?:[0-5][0-9])`,
	"seconds":         `(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)`,
	"time":            `((([01][0-9])|2[0-4]):(?:[0-5][0-9])(?::(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)))`,
	"iso8601Timezone": `(Z|([+-](([01][0-9])|2[0-4]):?([0-5][0-9])))`,
}

// deserializeStringList parses the legacy JSON-array-in-text column.
func deserializeStringList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}
