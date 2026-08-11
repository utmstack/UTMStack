package repository

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

type RuleBootstrap struct {
	srcDir string // image source rules (e.g. /utmstack/rules)
	store  *RuleStore
	db     *gorm.DB
}

func NewRuleBootstrap(srcDir string, store *RuleStore, db *gorm.DB) *RuleBootstrap {
	return &RuleBootstrap{srcDir: srcDir, store: store, db: db}
}

func (b *RuleBootstrap) Run(ctx context.Context) error {
	if err := b.seedSystemOverlay(); err != nil {
		_ = catcher.Error("eventprocessing: seeding system rules failed", err, nil)
	}
	return b.store.Load()
}

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
		_ = os.WriteFile(target, data, 0o644)
		return nil
	})
	if err != nil {
		return err
	}

	return b.pruneSystemOverlay(expected)
}

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
		if expected[rel] {
			return nil
		}
		if rmErr := os.Remove(path); rmErr != nil {
			_ = catcher.Error("eventprocessing: pruning orphaned system rule failed", rmErr,
				map[string]any{"rule": rel})
		}
		return nil
	})
}

func toRaw(s string) json.RawMessage {
	if s == "" {
		return nil
	}
	return json.RawMessage(s)
}
