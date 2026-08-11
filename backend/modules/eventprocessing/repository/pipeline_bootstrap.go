package repository

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type PipelineBootstrap struct {
	srcDir string
	store  *PipelineStore
	db     *gorm.DB
}

func NewPipelineBootstrap(srcDir string, store *PipelineStore, db *gorm.DB) *PipelineBootstrap {
	return &PipelineBootstrap{srcDir: srcDir, store: store, db: db}
}

func (b *PipelineBootstrap) Run(ctx context.Context) error {
	if err := b.seedSystemOverlay(); err != nil {
		_ = catcher.Error("eventprocessing: seeding system filters failed", err, nil)
	}
	return b.store.Load()
}

func (b *PipelineBootstrap) seedSystemOverlay() error {
	if _, err := os.Stat(b.srcDir); os.IsNotExist(err) {
		return nil
	}

	expected := make(map[string]bool)

	err := filepath.WalkDir(b.srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Ext(path) != PipelineFileExt {
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

var orderLinePattern = regexp.MustCompile(`(?m)^(\s*order:\s*)\d+(\s*#.*)?$`)

func preserveDestOrder(destPath string, newContent []byte) []byte {
	existing, err := os.ReadFile(destPath)
	if err != nil {
		return newContent
	}
	var existingCfg domain.PipelineSpec
	if err := yaml.Unmarshal(existing, &existingCfg); err != nil || len(existingCfg.Pipeline) == 0 {
		return newContent
	}
	order := existingCfg.Pipeline[0].Order
	if order == 0 {
		return newContent
	}
	return orderLinePattern.ReplaceAll(newContent, []byte(fmt.Sprintf("${1}%d$2", order)))
}

func (b *PipelineBootstrap) pruneSystemOverlay(expected map[string]bool) error {
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
