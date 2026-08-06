package usecase

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// FilterBootstrap seeds/refreshes the filter overlays at startup and, on the
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
	if err := b.seedSystemOverlay(); err != nil {
		_ = catcher.Error("eventprocessing: seeding system filters failed", err, nil)
	}
	return b.store.Load()
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
		destPath := target
		if _, err := os.Stat(disabled); err == nil {
			destPath = disabled
		}
		data = preserveDestOrder(destPath, data)
		_ = os.WriteFile(destPath, data, 0o644)
		return nil
	})
	if err != nil {
		return err
	}
	return b.pruneSystemOverlay(expected)
}

// orderLinePattern matches a pipeline entry's `order:` line, with an optional
// trailing comment, so preserveDestOrder can substitute just the number and
// leave everything else (comments, formatting) byte-for-byte untouched.
var orderLinePattern = regexp.MustCompile(`(?m)^(\s*order:\s*)\d+(\s*#.*)?$`)

// preserveDestOrder keeps whatever `order` already exists on disk at destPath
// (a customer/operator customization — including for system filters, whose
// order is editable even though their content isn't) instead of letting the
// freshly-shipped content clobber it on every boot. A destination that
// doesn't exist yet, or whose order is still the unset default (0), takes
// the shipped order as-is.
func preserveDestOrder(destPath string, newContent []byte) []byte {
	existing, err := os.ReadFile(destPath)
	if err != nil {
		return newContent
	}
	var existingCfg filterConfig
	if err := yaml.Unmarshal(existing, &existingCfg); err != nil || len(existingCfg.Pipeline) == 0 {
		return newContent
	}
	order := existingCfg.Pipeline[0].Order
	if order == 0 {
		return newContent
	}
	return orderLinePattern.ReplaceAll(newContent, []byte(fmt.Sprintf("${1}%d$2", order)))
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
