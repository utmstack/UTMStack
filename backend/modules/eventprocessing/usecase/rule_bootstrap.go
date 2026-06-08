package usecase

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"gorm.io/gorm"
)

// RuleBootstrap prepares the rule overlays at startup: it seeds the read-only
// system overlay from the image's source rules and, on an in-place upgrade,
// migrates the user's rules out of the legacy DB into the user overlay.
type RuleBootstrap struct {
	srcDir string // image source rules (e.g. /utmstack/rules)
	store  *RuleStore
	db     *gorm.DB
}

func NewRuleBootstrap(srcDir string, store *RuleStore, db *gorm.DB) *RuleBootstrap {
	return &RuleBootstrap{srcDir: srcDir, store: store, db: db}
}

// Run is idempotent and safe to call on every boot.
func (b *RuleBootstrap) Run(ctx context.Context) error {
	// 0. Remove anything sitting loose in the rules root. An older event
	//    processor wrote rule files flat into /workdir/rules; from now on rules
	//    live ONLY under the system/ and user/ overlays. No-op once clean.
	if err := b.purgeLooseRules(); err != nil {
		_ = catcher.Error("eventprocessing: purging loose rules failed", err, nil)
	}

	// 1. Seed/refresh the system overlay from the image source rules. A system
	//    rule the operator has disabled (a .disabled file already present) is
	//    left untouched so the disable survives upgrades.
	if err := b.seedSystemOverlay(); err != nil {
		_ = catcher.Error("eventprocessing: seeding system rules failed", err, nil)
	}

	// 2. Make the in-memory store reflect the overlays before migrating, so the
	//    legacy migration can resolve system rules by name.
	if err := b.store.Load(); err != nil {
		return err
	}

	// 3. One-time migration from the legacy DB (in-place upgrade from the
	//    Java/DB stack). Idempotent: user rules already present are skipped.
	//    Once every rule is carried over to YAML the legacy tables are dropped
	//    (utm_correlation_rules + its join); a partial failure keeps them for a
	//    retry on the next boot.
	if b.db != nil {
		if err := b.migrateLegacyRules(ctx); err != nil {
			_ = catcher.Error("eventprocessing: legacy rule migration failed", err, nil)
		}
		// Reload to pick up migrated user rules / disabled-state changes.
		if err := b.store.Load(); err != nil {
			return err
		}
	}

	return nil
}

// purgeLooseRules deletes every entry directly under the rules root except the
// system/ and user/ overlays, so rules never sit loose in /workdir/rules.
// Idempotent: a no-op once only system/ and user/ remain.
func (b *RuleBootstrap) purgeLooseRules() error {
	root := filepath.Dir(b.store.systemDir)
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == SystemSubdir || e.Name() == UserSubdir {
			continue // keep the overlays
		}
		if err := os.RemoveAll(filepath.Join(root, e.Name())); err != nil {
			_ = catcher.Error("eventprocessing: removing loose rule entry failed", err,
				map[string]any{"entry": e.Name()})
		}
	}
	return nil
}

// seedSystemOverlay makes the system overlay mirror the image source rules: it
// copies/refreshes every source rule (canonical .yaml extension, preserving an
// operator's .disabled marker) and prunes system rules whose source was removed
// from the image, so a deprecated rule never lingers (legacy: cleanupOrphanedRules).
func (b *RuleBootstrap) seedSystemOverlay() error {
	if _, err := os.Stat(b.srcDir); os.IsNotExist(err) {
		return nil
	}

	expected := make(map[string]bool) // canonical relPaths present in the image source

	err := filepath.WalkDir(b.srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != RuleFileExt {
			return nil
		}

		rel, err := filepath.Rel(b.srcDir, path)
		if err != nil {
			return nil
		}
		expected[rel] = true
		target := filepath.Join(b.store.systemDir, rel)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil
		}

		disabled := target + DisabledSuffix
		if _, err := os.Stat(disabled); err == nil {
			_ = os.WriteFile(disabled, data, 0o644)
		} else {
			_ = os.WriteFile(target, data, 0o644)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return b.pruneSystemOverlay(expected)
}

// pruneSystemOverlay removes any system overlay file (enabled or .disabled) whose
// canonical rule is no longer shipped in the image source.
func (b *RuleBootstrap) pruneSystemOverlay(expected map[string]bool) error {
	if _, err := os.Stat(b.store.systemDir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(b.store.systemDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(b.store.systemDir, path)
		if relErr != nil {
			return nil
		}
		// Canonicalize: a disabled file X.yaml.disabled maps to source X.yaml.
		canon := strings.TrimSuffix(rel, DisabledSuffix)
		if expected[canon] {
			return nil
		}
		if rmErr := os.Remove(path); rmErr != nil {
			_ = catcher.Error("eventprocessing: pruning orphaned system rule failed", rmErr,
				map[string]any{"rule": rel})
		}
		return nil
	})
}

// migrateLegacyRules moves user-owned rules from utm_correlation_rules into the
// user overlay and reconciles disabled system rules. It is idempotent.
func (b *RuleBootstrap) migrateLegacyRules(ctx context.Context) error {
	// The table only exists on an in-place upgrade from the Java/DB stack. On a
	// fresh install it is never created (rules are YAML-direct), so there is
	// nothing to migrate.
	if !b.db.Migrator().HasTable(&domain.UtmCorrelationRules{}) {
		return nil
	}

	var legacy []domain.UtmCorrelationRules
	if err := b.db.WithContext(ctx).Preload("DataTypes").Find(&legacy).Error; err != nil {
		return err
	}
	if len(legacy) == 0 {
		return nil // nothing to migrate (fresh install or already migrated)
	}

	// Map system rule name -> relPath so a disabled legacy system rule can be
	// reconciled against its overlay file.
	systemByName := make(map[string]string)
	b.store.mu.RLock()
	for _, sr := range b.store.rules {
		if sr.system {
			systemByName[sr.Name] = sr.RelPath
		}
	}
	b.store.mu.RUnlock()

	failed := 0
	for i := range legacy {
		row := &legacy[i]

		if row.SystemOwner {
			// System rules ship as files; only carry over a disable.
			if !row.RuleActive {
				if relPath, ok := systemByName[row.RuleName]; ok {
					if err := b.store.SetEnabled(relPath, false); err != nil {
						_ = catcher.Error("eventprocessing: reconciling disabled system rule failed", err,
							map[string]any{"rule": row.RuleName})
						failed++
					}
				}
			}
			continue
		}

		// User rule: recreate it in the user overlay (skip if already present).
		rule := legacyToRule(row)
		created, err := b.store.Create(rule)
		if err != nil {
			if os.IsExist(err) {
				continue // already migrated
			}
			_ = catcher.Error("eventprocessing: migrating legacy user rule failed", err,
				map[string]any{"rule": row.RuleName})
			failed++
			continue
		}
		if !row.RuleActive {
			if err := b.store.SetEnabled(created.RelPath, false); err != nil {
				_ = catcher.Error("eventprocessing: disabling migrated user rule failed", err,
					map[string]any{"rule": row.RuleName})
				failed++
			}
		}
	}

	// Drop the legacy tables only after a fully-clean migration, so a partial
	// failure is retried (idempotently) on the next boot instead of losing rows.
	if failed > 0 {
		catcher.Warn("eventprocessing: legacy rule migration had failures — keeping utm_correlation_rules for retry on next boot", nil)
		return nil
	}
	return b.dropLegacyRuleTables()
}

// dropLegacyRuleTables removes the migrated-away DB tables once every rule has
// been carried over to the YAML overlay. Idempotent (IF EXISTS) and safe to
// re-run. The join table is dropped first as it FK-references the rules table.
// utm_data_types is dropped too: it was read here (for rule data-type names) and
// by the integrations pre-migration (to backfill utm_module.data_type); after
// both consume it, data types live as utm_module.data_type and the table is dead.
func (b *RuleBootstrap) dropLegacyRuleTables() error {
	if err := b.db.Exec("DROP TABLE IF EXISTS utm_group_rules_data_type CASCADE").Error; err != nil {
		return err
	}
	if err := b.db.Exec("DROP TABLE IF EXISTS utm_correlation_rules CASCADE").Error; err != nil {
		return err
	}
	if err := b.db.Exec("DROP TABLE IF EXISTS utm_data_types CASCADE").Error; err != nil {
		return err
	}
	catcher.Info("eventprocessing: legacy correlation-rule and data-type tables migrated and dropped", nil)
	return nil
}

// legacyToRule converts a legacy DB row into the on-disk Rule shape, reusing the
// same raw-JSON decoders the usecase uses for request payloads.
func legacyToRule(row *domain.UtmCorrelationRules) Rule {
	names := make([]string, 0, len(row.DataTypes))
	for _, dt := range row.DataTypes {
		names = append(names, dt.DataType)
	}
	return Rule{
		Name:          row.RuleName,
		Adversary:     row.RuleAdversary,
		Category:      row.RuleCategory,
		Technique:     row.RuleTechnique,
		Description:   row.RuleDescription,
		DataTypes:     names,
		Impact:        Impact{Confidentiality: row.RuleConfidentiality, Integrity: row.RuleIntegrity, Availability: row.RuleAvailability},
		Where:         rawToWhere(toRaw(row.RuleDefinitionDef)),
		References:    rawToAnySlice(toRaw(row.RuleReferencesDef)),
		AfterEvents:   rawToAny(toRaw(row.AfterEventsDef)),
		GroupBy:       rawToStrSlice(toRaw(row.RuleGroupByDef)),
		DeduplicateBy: rawToStrSlice(toRaw(row.DeduplicateByDef)),
	}
}

func toRaw(s string) json.RawMessage {
	if s == "" {
		return nil
	}
	return json.RawMessage(s)
}
