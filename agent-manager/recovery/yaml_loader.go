package recovery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// rawFile holds the unparsed content of a single YAML file read from disk.
type rawFile struct {
	Path    string // absolute path, used in error messages
	Content []byte
	Hash    string // SHA-256 hex of Content, computed once at read time
}

// ParseErr describes a single per-file failure (parse or read). Non-fatal — callers
// log these and continue processing the remaining files.
type ParseErr struct {
	Path string
	Err  error
}

// ReadDir lists all *.yaml and *.yml files under path, reads their contents,
// and computes SHA-256 hashes. A missing directory returns (nil, nil) with no
// error — recovery is optional and agent-manager must boot even with no files
// present (design decision D1).
//
// Per-file read errors (permission, IO) are returned as nil files + non-nil error
// only when the directory itself cannot be opened. Individual file failures are
// aggregated in the returned ParseErr slice (non-fatal, caller logs them).
func ReadDir(path string) ([]rawFile, []ParseErr) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing directory is NOT an error — recovery YAML dir is optional.
			return nil, nil
		}
		return nil, []ParseErr{{Path: path, Err: fmt.Errorf("open recovery dir: %w", err)}}
	}

	var files []rawFile
	var errs []ParseErr

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".yaml") && !strings.HasSuffix(lower, ".yml") {
			continue
		}

		absPath := filepath.Join(path, name)
		content, readErr := os.ReadFile(absPath)
		if readErr != nil {
			errs = append(errs, ParseErr{
				Path: absPath,
				Err:  fmt.Errorf("read file: %w", readErr),
			})
			continue
		}

		sum := sha256.Sum256(content)
		files = append(files, rawFile{
			Path:    absPath,
			Content: content,
			Hash:    hex.EncodeToString(sum[:]),
		})
	}

	return files, errs
}

// ParseAll unmarshals each rawFile into a RecoveryYAML using goccy/go-yaml.
// Per-file unmarshal errors are appended to the returned ParseErr slice (non-fatal).
// Only successfully parsed files appear in the first return value.
// The parsed struct's Hash field is populated from rawFile.Hash so downstream
// code always knows which file a given RecoveryYAML originated from.
func ParseAll(files []rawFile) ([]RecoveryYAML, []ParseErr) {
	out := make([]RecoveryYAML, 0, len(files))
	var errs []ParseErr

	for _, f := range files {
		var r RecoveryYAML
		if err := yaml.Unmarshal(f.Content, &r); err != nil {
			errs = append(errs, ParseErr{
				Path: f.Path,
				Err:  fmt.Errorf("unmarshal yaml: %w", err),
			})
			continue
		}
		r.Hash = f.Hash
		out = append(out, r)
	}

	return out, errs
}
