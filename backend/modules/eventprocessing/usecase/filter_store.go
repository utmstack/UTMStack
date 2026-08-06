package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"
)

// FilterEntry is an in-memory record of a filter file.
type FilterEntry struct {
	RelPath string // canonical relPath (e.g. "syslog/syslog-generic.yaml")
	Content []byte // raw YAML content (pipeline: format)
	System  bool   // true → lives in systemDir; false → userDir
	Active  bool   // false when a .disabled sibling exists
	// TenantId scopes this filter to one tenant — read from the content's own
	// `pipeline[].tenantId` field, the same one the engine reads. Empty means
	// global/system, same as before this field existed.
	TenantId string
}

// genericPipelineFile is a loosely-typed mirror of the `pipeline:` YAML
// format, used only to read/write the tenantId field without needing to
// model the entire Step schema (grok, kv, csv, ...) in this package.
type genericPipelineFile struct {
	Pipeline []map[string]any `yaml:"pipeline"`
}

// filterTenantID reads the tenantId of a filter's first (and, in practice,
// only) pipeline entry — mirrors filter.go's firstPipelineOrder.
func filterTenantID(content []byte) string {
	var cfg genericPipelineFile
	if err := yaml.Unmarshal(content, &cfg); err != nil || len(cfg.Pipeline) == 0 {
		return ""
	}
	v, _ := cfg.Pipeline[0]["tenantId"].(string)
	return v
}

// withTenantID sets (or clears) tenantId on every pipeline entry in content.
func withTenantID(content []byte, tenantId string) ([]byte, error) {
	var cfg genericPipelineFile
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	for i := range cfg.Pipeline {
		if tenantId == "" {
			delete(cfg.Pipeline[i], "tenantId")
		} else {
			cfg.Pipeline[i]["tenantId"] = tenantId
		}
	}
	return yaml.Marshal(cfg)
}

// FilterStore manages two overlay directories for pipeline-format filter YAML
// files, following the same pattern as RuleStore:
//
//	systemDir/ — read-only for the operator (shipped with the image)
//	userDir/   — user-created and user-editable filters
type FilterStore struct {
	systemDir string
	userDir   string
	mu        sync.RWMutex
	filters   map[string]*FilterEntry // keyed by RelPath
}

func NewFilterStore(systemDir, userDir string) *FilterStore {
	return &FilterStore{
		systemDir: systemDir,
		userDir:   userDir,
		filters:   make(map[string]*FilterEntry),
	}
}

// Load rescans both overlays and rebuilds the in-memory map.
func (s *FilterStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filters = make(map[string]*FilterEntry)

	for _, root := range []struct {
		dir    string
		system bool
	}{{s.systemDir, true}, {s.userDir, false}} {
		if _, err := os.Stat(root.dir); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(root.dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			base := path
			active := true
			if strings.HasSuffix(base, DisabledSuffix) {
				base = strings.TrimSuffix(base, DisabledSuffix)
				active = false
			}
			if filepath.Ext(base) != FilterFileExt {
				return nil
			}
			rel, err := filepath.Rel(root.dir, base)
			if err != nil {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			s.filters[rel] = &FilterEntry{
				RelPath:  rel,
				Content:  data,
				System:   root.system,
				Active:   active,
				TenantId: filterTenantID(data),
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Watch periodically reloads the filter overlays so this replica picks up
// filters/enable-state written by another backend replica sharing the same
// directories. Intended to run as its own goroutine for the lifetime of ctx.
func (s *FilterStore) Watch(ctx context.Context) {
	ticker := time.NewTicker(reloadInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Load(); err != nil {
				_ = catcher.Error("eventprocessing: periodic filter reload failed", err, nil)
			}
		}
	}
}

// List returns all filter entries (copy).
func (s *FilterStore) List() []FilterEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]FilterEntry, 0, len(s.filters))
	for _, f := range s.filters {
		out = append(out, *f)
	}
	return out
}

// GetByRelPath returns the entry for the given relPath, or nil.
func (s *FilterStore) GetByRelPath(relPath string) *FilterEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f := s.filters[relPath]
	if f == nil {
		return nil
	}
	cp := *f
	return &cp
}

// Create writes a new filter file to userDir. tenantId scopes the filter to
// one tenant — empty means a global user filter, placed exactly as before
// this field existed. A tenant-owned filter lives in its own subfolder
// (userDir/<tenantId>/), and its filename is suffixed with the tenant id too:
// identity (see pipelineIdentity in go-sdk) is basename-only, so without the
// suffix two tenants picking the same filter name would collide on the exact
// same identity the engine uses. Returns os.ErrExist if already present.
func (s *FilterStore) Create(relPath string, content []byte, tenantId string) (*FilterEntry, error) {
	if tenantId != "" {
		injected, err := withTenantID(content, tenantId)
		if err != nil {
			return nil, fmt.Errorf("invalid filter content: %w", err)
		}
		content = injected

		base := filepath.Base(relPath)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		relPath = filepath.ToSlash(filepath.Join(tenantId, name+"-"+tenantId+ext))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.filters[relPath]; exists {
		return nil, os.ErrExist
	}
	target := filepath.Join(s.userDir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return nil, err
	}
	if err := atomicWrite(target, content); err != nil {
		return nil, err
	}
	entry := &FilterEntry{RelPath: relPath, Content: content, System: false, Active: true, TenantId: tenantId}
	s.filters[relPath] = entry
	cp := *entry
	return &cp, nil
}

// Update overwrites an existing user filter. Returns ErrFilterSystemOwner for system filters.
func (s *FilterStore) Update(relPath string, content []byte) (*FilterEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.filters[relPath]
	if !ok {
		return nil, ErrFilterNotFound
	}
	if existing.System {
		return nil, ErrFilterSystemOwner
	}
	target := filepath.Join(s.userDir, relPath)
	if err := atomicWrite(target, content); err != nil {
		return nil, err
	}
	existing.Content = content
	cp := *existing
	return &cp, nil
}

// UpdateOrder overwrites a filter's content in whichever overlay (system or
// user) currently owns it. Unlike Update, it does NOT reject system filters —
// their order is customer-controlled even though their steps/content stay
// read-only via Update/Delete.
func (s *FilterStore) UpdateOrder(relPath string, content []byte) (*FilterEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.filters[relPath]
	if !ok {
		return nil, ErrFilterNotFound
	}
	dir := s.userDir
	if existing.System {
		dir = s.systemDir
	}
	target := filepath.Join(dir, relPath)
	if err := atomicWrite(target, content); err != nil {
		return nil, err
	}
	existing.Content = content
	cp := *existing
	return &cp, nil
}

// Delete removes a user filter file. Returns ErrFilterSystemOwner for system filters.
func (s *FilterStore) Delete(relPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.filters[relPath]
	if !ok {
		return ErrFilterNotFound
	}
	if existing.System {
		return ErrFilterSystemOwner
	}
	target := filepath.Join(s.userDir, relPath)
	// Try enabled path; fall back to disabled.
	if err := os.Remove(target); err != nil {
		_ = os.Remove(target + DisabledSuffix)
	}
	delete(s.filters, relPath)
	return nil
}

// SetEnabled renames .yaml ↔ .yaml.disabled.
func (s *FilterStore) SetEnabled(relPath string, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.filters[relPath]
	if !ok {
		return ErrFilterNotFound
	}
	dir := s.userDir
	if existing.System {
		dir = s.systemDir
	}
	enabled := filepath.Join(dir, relPath)
	disabled := enabled + DisabledSuffix
	if active {
		if _, err := os.Stat(disabled); err == nil {
			if err := os.Rename(disabled, enabled); err != nil {
				return err
			}
		}
	} else {
		if _, err := os.Stat(enabled); err == nil {
			if err := os.Rename(enabled, disabled); err != nil {
				return err
			}
		}
	}
	existing.Active = active
	return nil
}

// atomicWrite writes data to path via a temp file + rename (0600).
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ErrFilterNotFound and ErrFilterSystemOwner are sentinel errors for the store.
// They are intentionally separate from the domain errors so the store can be
// used without importing the domain package.
var (
	ErrFilterNotFound    = errFilterNotFound{}
	ErrFilterSystemOwner = errFilterSystemOwner{}
)

type errFilterNotFound struct{}
type errFilterSystemOwner struct{}

func (errFilterNotFound) Error() string    { return "filter not found" }
func (errFilterSystemOwner) Error() string { return "system filter cannot be modified" }
