package usecase

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"gorm.io/gorm"
)

// FilterBootstrap seeds/refreshes the filter overlays at startup and, on the
// first boot after an in-place upgrade, migrates legacy DB filters to the user
// overlay before dropping the table.
type FilterBootstrap struct {
	srcDir string
	store  *FilterStore
	db     *gorm.DB
}

func NewFilterBootstrap(srcDir string, store *FilterStore, db *gorm.DB) *FilterBootstrap {
	return &FilterBootstrap{srcDir: srcDir, store: store, db: db}
}

// Run is idempotent and safe to call on every boot.
func (b *FilterBootstrap) Run(ctx context.Context) error {
	// 0. Remove anything sitting loose in the filters root (left by the legacy
	//    plugins/config which wrote flat files there).
	if err := b.purgeLooseFilters(); err != nil {
		_ = catcher.Error("eventprocessing: purging loose filter files failed", err, nil)
	}

	// 1. Mirror the image source into the system overlay (+ prune orphans).
	if err := b.seedSystemOverlay(); err != nil {
		_ = catcher.Error("eventprocessing: seeding system filter overlay failed", err, nil)
	}

	// 2. Load overlays into memory.
	if err := b.store.Load(); err != nil {
		return err
	}

	// 3. One-time migration from the legacy DB → user overlay, then drop.
	if b.db != nil {
		if err := b.migrateLegacyFilters(ctx); err != nil {
			_ = catcher.Error("eventprocessing: legacy filter migration failed", err, nil)
		}
		if err := b.store.Load(); err != nil {
			return err
		}
	}

	return nil
}

func (b *FilterBootstrap) purgeLooseFilters() error {
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
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, e.Name())); err != nil {
			_ = catcher.Error("eventprocessing: removing loose filter entry failed", err,
				map[string]any{"entry": e.Name()})
		}
	}
	return nil
}

// seedSystemOverlay mirrors /utmstack/filters → system/ and prunes deleted ones.
func (b *FilterBootstrap) seedSystemOverlay() error {
	if _, err := os.Stat(b.srcDir); os.IsNotExist(err) {
		return nil
	}

	expected := make(map[string]bool)

	err := filepath.WalkDir(b.srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Ext(path) != FilterFileExt {
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
		// Preserve the operator's disabled state but refresh the content into
		// whichever file exists (so updates reach disabled filters too).
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

func (b *FilterBootstrap) pruneSystemOverlay(expected map[string]bool) error {
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
		canon := strings.TrimSuffix(rel, DisabledSuffix)
		if expected[canon] {
			return nil
		}
		if rmErr := os.Remove(path); rmErr != nil {
			_ = catcher.Error("eventprocessing: pruning orphaned system filter failed", rmErr,
				map[string]any{"filter": rel})
		}
		return nil
	})
}

// migrateLegacyFilters reads the legacy utm_logstash_filter table, wraps each
// row's raw content in a pipeline: YAML envelope, writes it to the user overlay,
// then drops the table once all rows are processed without error.
func (b *FilterBootstrap) migrateLegacyFilters(ctx context.Context) error {
	if !b.db.Migrator().HasTable(&domain.UtmLogstashFilter{}) {
		return nil
	}

	var legacy []domain.UtmLogstashFilter
	if err := b.db.WithContext(ctx).Find(&legacy).Error; err != nil {
		return err
	}
	if len(legacy) == 0 {
		return b.dropLegacyFilterTable()
	}

	dataTypeByID := b.loadDataTypeNames(ctx)

	failed := 0
	for i := range legacy {
		row := &legacy[i]
		if !row.IsActive || row.SystemOwner {
			continue
		}
		relPath := fmt.Sprintf("legacy/%d.yaml", row.ID)
		var dataType string
		if row.DataTypeID != nil {
			dataType = dataTypeByID[*row.DataTypeID]
		}
		content := wrapInPipelineYAML(row.LogstashFilter, dataType)

		if _, err := b.store.Create(relPath, content); err != nil {
			if os.IsExist(err) {
				continue
			}
			_ = catcher.Error("eventprocessing: migrating legacy filter failed", err,
				map[string]any{"id": row.ID, "name": row.FilterName})
			failed++
		}
	}

	if failed > 0 {
		catcher.Warn("eventprocessing: legacy filter migration had failures — keeping utm_logstash_filter for retry", nil)
		return nil
	}
	return b.dropLegacyFilterTable()
}

func (b *FilterBootstrap) dropLegacyFilterTable() error {
	if err := b.db.Exec("DROP TABLE IF EXISTS utm_logstash_filter CASCADE").Error; err != nil {
		return err
	}
	catcher.Info("eventprocessing: legacy utm_logstash_filter migrated to YAML and dropped", nil)

	if !b.db.Migrator().HasTable("utm_correlation_rules") {
		if err := b.db.Exec("DROP TABLE IF EXISTS utm_data_types CASCADE").Error; err != nil {
			return err
		}
	}
	return nil
}

func (b *FilterBootstrap) loadDataTypeNames(ctx context.Context) map[int64]string {
	out := map[int64]string{}
	if !b.db.Migrator().HasTable(&domain.UtmDataTypes{}) {
		return out
	}
	var rows []domain.UtmDataTypes
	if err := b.db.WithContext(ctx).Find(&rows).Error; err != nil {
		_ = catcher.Error("eventprocessing: reading utm_data_types for filter migration failed", err, nil)
		return out
	}
	for i := range rows {
		if rows[i].DataType != "" {
			out[rows[i].ID] = rows[i].DataType
		}
	}
	return out
}

func wrapInPipelineYAML(rawDSL, dataType string) []byte {

	lines := strings.Split(rawDSL, "\n")
	for i, l := range lines {
		lines[i] = "            " + l
	}
	indented := strings.Join(lines, "\n")

	const header = "# migrated from legacy utm_logstash_filter\npipeline:\n"
	const stepsBlock = "    steps:\n      - logstash:\n          filter: |\n%s\n"

	// This lands in the user overlay, so it needs an explicit order at or
	// above customFilterDefaultOrder — otherwise it would default to 0 and
	// could tie with a system filter sharing the same dataType.
	if dataType != "" {
		return []byte(fmt.Sprintf(header+"  - dataType: %s\n    order: %d\n"+stepsBlock, dataType, customFilterDefaultOrder, indented))
	}
	return []byte(fmt.Sprintf(header+"  - order: %d\n    steps:\n      - logstash:\n          filter: |\n%s\n", customFilterDefaultOrder, indented))
}
