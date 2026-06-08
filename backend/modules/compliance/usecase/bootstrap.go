package usecase

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SeedSystemOverlay mirrors the vendor source directory (shipped in the image)
// into the system overlay: it copies every .yaml, preserves an operator-disabled
// system file (a .disabled sibling), and prunes system files no longer shipped.
// The user overlay is never touched. Mirrors the rule/filter bootstrap.
func SeedSystemOverlay(srcDir, systemDir string) error {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}
	expected := map[string]bool{}
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Ext(path) != fileExt {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}
		expected[filepath.ToSlash(rel)] = true
		target := filepath.Join(systemDir, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil
		}
		// Preserve the operator's disabled state but refresh the content into
		// whichever file exists (so updates reach disabled files too).
		disabled := target + disabledSuffix
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
	return pruneSystemOverlay(systemDir, expected)
}

func pruneSystemOverlay(systemDir string, expected map[string]bool) error {
	if _, err := os.Stat(systemDir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(systemDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(systemDir, path)
		if relErr != nil {
			return nil
		}
		canon := strings.TrimSuffix(filepath.ToSlash(rel), disabledSuffix)
		if expected[canon] {
			return nil
		}
		_ = os.Remove(path)
		return nil
	})
}
