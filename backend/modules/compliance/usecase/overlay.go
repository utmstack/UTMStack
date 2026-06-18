package usecase

import (
	"os"
	"path/filepath"
	"strings"
)

// Shared system/user overlay mechanics for the file-backed control library and
// framework checklists, mirroring eventprocessing's rule/filter stores: system =
// vendor (read-only, refreshed on upgrade), user = created/cloned (never
// overwritten), `.disabled` filename suffix toggles enabled state.

const (
	SystemSubdir   = "system"
	UserSubdir     = "user"
	disabledSuffix = ".disabled"
	fileExt        = ".yaml"
)

type scannedFile struct {
	relPath string
	data    []byte
	enabled bool
	system  bool
}

// safeID rejects ids/keys that are empty or could escape the overlay directory
// (path separators / traversal), guarding the user-overlay writes.
func safeID(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	return !strings.ContainsAny(id, "/\\") && !strings.Contains(id, "..")
}

// atomicWrite writes data to path via a temp file + rename (creating parents).
func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// setEnabledFile toggles relPath ↔ relPath.disabled within dir.
func setEnabledFile(dir, relPath string, enabled bool) error {
	enabledPath := filepath.Join(dir, relPath)
	disabledPath := enabledPath + disabledSuffix
	if enabled {
		if _, err := os.Stat(disabledPath); err == nil {
			return os.Rename(disabledPath, enabledPath)
		}
		return nil
	}
	if _, err := os.Stat(enabledPath); err == nil {
		return os.Rename(enabledPath, disabledPath)
	}
	return nil
}

// scanYAML walks dir recursively and returns every .yaml (and .yaml.disabled)
// file. relPath is canonical (without the .disabled suffix).
func scanYAML(dir string, system bool) ([]scannedFile, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	var out []scannedFile
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		base := path
		enabled := true
		if strings.HasSuffix(base, disabledSuffix) {
			base = strings.TrimSuffix(base, disabledSuffix)
			enabled = false
		}
		if filepath.Ext(base) != fileExt {
			return nil
		}
		rel, err := filepath.Rel(dir, base)
		if err != nil {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		out = append(out, scannedFile{relPath: filepath.ToSlash(rel), data: data, enabled: enabled, system: system})
		return nil
	})
	return out, err
}
