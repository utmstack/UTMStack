package usecase

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/threatwinds/go-sdk/catcher"
)

func (s *FlowStore) SeedSystem(srcDir string) error {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}

	shipped := make(map[string]bool)

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != FlowFileExt {
			return nil
		}

		rel, relErr := filepath.Rel(srcDir, path)
		if relErr != nil {
			return nil
		}
		shipped[rel] = true

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			_ = catcher.Error("soar: reading a shipped flow failed", readErr, map[string]any{"flow": rel})
			return nil
		}

		target := filepath.Join(s.systemDir, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			_ = catcher.Error("soar: preparing the system overlay failed", err, map[string]any{"flow": rel})
			return nil
		}
		if err := writeFileAtomic(target, data); err != nil {
			_ = catcher.Error("soar: writing a shipped flow failed", err, map[string]any{"flow": rel})
		}
		return nil
	})
	if err != nil {
		return err
	}

	return s.pruneSystem(shipped)
}

func (s *FlowStore) pruneSystem(shipped map[string]bool) error {
	if _, err := os.Stat(s.systemDir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(s.systemDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(s.systemDir, path)
		if relErr != nil || shipped[rel] {
			return nil
		}
		if rmErr := os.Remove(path); rmErr != nil {
			_ = catcher.Error("soar: pruning a retired system flow failed", rmErr, map[string]any{"flow": rel})
		}
		return nil
	})
}

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
