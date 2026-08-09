package usecase

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

// pipelineWriter is the writer for tenants.yaml and patterns.yaml in the
// pipeline directory shared with the event processor. The engine reads these
// via loadCfg (Walk + ReadPbYaml → protojson.Unmarshal), so the YAML field
// names must be the protobuf JSON names (camelCase as defined in
// plugins.proto).
//
// Every write acquires an exclusive file lock, re-reads the current file,
// applies its one change, and writes back atomically — there is no
// long-lived in-memory "authoritative" state, so concurrent backend replicas
// sharing this directory can never clobber each other's writes (the
// lost-update this replaced: two independent mutators — asset CIA config and
// rule/pipeline enable state — each holding their own stale in-memory copy).
type pipelineWriter struct {
	dir string
	mu  sync.Mutex // in-process fast path; withFileLock handles cross-process safety
}

func NewPipelineWriter(dir string) *pipelineWriter {
	return &pipelineWriter{dir: dir}
}

// --- YAML structs (proto JSON field names) ---

type tenantFileYAML struct {
	Tenants []tenantEntryYAML `yaml:"tenants"`
}

type tenantEntryYAML struct {
	ID                string          `yaml:"id"`
	Name              string          `yaml:"name"`
	Assets            []assetFileYAML `yaml:"assets,omitempty"`
	DisabledRules     []string        `yaml:"disabledRules,omitempty"`
	DisabledPipelines []string        `yaml:"disabledPipelines,omitempty"`
}

type TenantAssets struct {
	ID     string
	Name   string
	Assets []assetFileYAML
}

type assetFileYAML struct {
	Name            string   `yaml:"name"`
	Hostnames       []string `yaml:"hostnames,omitempty"`
	Ips             []string `yaml:"ips,omitempty"`
	Confidentiality uint32   `yaml:"confidentiality"`
	Availability    uint32   `yaml:"availability"`
	Integrity       uint32   `yaml:"integrity"`
}

type patternsFileYAML struct {
	Patterns map[string]string `yaml:"patterns"`
}

func (w *pipelineWriter) readTenantFile() (tenantFileYAML, error) {
	data, err := os.ReadFile(filepath.Join(w.dir, TenantFileName))
	if os.IsNotExist(err) {
		return tenantFileYAML{}, nil
	}
	if err != nil {
		return tenantFileYAML{}, err
	}
	var content tenantFileYAML
	if err := yaml.Unmarshal(data, &content); err != nil {
		return tenantFileYAML{}, err
	}
	return content, nil
}

// readTenantEntry is readTenantFile for the callers that only care about the
// opt-out lists. Those are an install-wide setting replicated onto every tenant
// entry, so reading the first is reading all of them.
func (w *pipelineWriter) readTenantEntry() (tenantEntryYAML, error) {
	content, err := w.readTenantFile()
	if err != nil {
		return tenantEntryYAML{}, err
	}
	if len(content.Tenants) == 0 {
		return tenantEntryYAML{ID: DefaultTenantID, Name: DefaultTenantName}, nil
	}
	return content.Tenants[0], nil
}

// writeTenantsLocked writes one entry per tenant, carrying the install-wide
// opt-out lists onto each. Caller must hold both w.mu and the file lock.
//
// The default tenant is always present even with no assets of its own: it is
// where a single-tenant install lives, and dropping it would take the opt-out
// lists with it.
func (w *pipelineWriter) writeTenantsLocked(tenants []TenantAssets, disabledRules, disabledPipelines []string) error {
	sort.Strings(disabledRules)
	sort.Strings(disabledPipelines)

	out := make([]tenantEntryYAML, 0, len(tenants)+1)
	seenDefault := false
	for _, t := range tenants {
		if t.ID == "" {
			continue
		}
		if t.ID == DefaultTenantID {
			seenDefault = true
			if t.Name == "" {
				t.Name = DefaultTenantName
			}
		}
		out = append(out, tenantEntryYAML{
			ID:                t.ID,
			Name:              t.Name,
			Assets:            t.Assets,
			DisabledRules:     disabledRules,
			DisabledPipelines: disabledPipelines,
		})
	}
	if !seenDefault {
		out = append(out, tenantEntryYAML{
			ID:                DefaultTenantID,
			Name:              DefaultTenantName,
			DisabledRules:     disabledRules,
			DisabledPipelines: disabledPipelines,
		})
	}
	// Ordered so an unchanged projection produces a byte-identical file and
	// does not look like a change to anything watching it.
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	return w.writeYAML(TenantFileName, tenantFileYAML{Tenants: out})
}

// writeTenantEntryLocked rewrites the opt-out lists across every tenant already
// on disk, leaving each one's assets alone.
func (w *pipelineWriter) writeTenantEntryLocked(t tenantEntryYAML) error {
	content, err := w.readTenantFile()
	if err != nil {
		return err
	}
	existing := make([]TenantAssets, 0, len(content.Tenants))
	for _, e := range content.Tenants {
		existing = append(existing, TenantAssets{ID: e.ID, Name: e.Name, Assets: e.Assets})
	}
	return w.writeTenantsLocked(existing, t.DisabledRules, t.DisabledPipelines)
}

// WriteTenants atomically rewrites tenants.yaml with one entry per tenant,
// keeping whatever disabled-rules/disabled-pipelines state is already on disk.
//
// The assets are replaced wholesale rather than merged: the caller recomputes
// them from the database every time, so what it passes is the whole truth and
// a merge would resurrect assets that were deleted.
func (w *pipelineWriter) WriteTenants(tenants []TenantAssets) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return withFileLock(w.tenantLockPath(), func() error {
		cur, err := w.readTenantEntry()
		if err != nil {
			return err
		}
		return w.writeTenantsLocked(tenants, cur.DisabledRules, cur.DisabledPipelines)
	})
}

// SetRuleDisabled records a rule (identified the same way the event processor
// does: filename without directory or extension) as disabled/enabled and
// persists it to tenants.yaml, leaving everything else untouched.
func (w *pipelineWriter) SetRuleDisabled(identity string, disabled bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return withFileLock(w.tenantLockPath(), func() error {
		t, err := w.readTenantEntry()
		if err != nil {
			return err
		}
		t.DisabledRules = toggleIdentity(t.DisabledRules, identity, disabled)
		return w.writeTenantEntryLocked(t)
	})
}

// SetPipelineDisabled is the filter/pipeline equivalent of SetRuleDisabled —
// same tenants.yaml, same opt-out mechanism, separate list.
func (w *pipelineWriter) SetPipelineDisabled(identity string, disabled bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return withFileLock(w.tenantLockPath(), func() error {
		t, err := w.readTenantEntry()
		if err != nil {
			return err
		}
		t.DisabledPipelines = toggleIdentity(t.DisabledPipelines, identity, disabled)
		return w.writeTenantEntryLocked(t)
	})
}

// PokeRuleReload bumps a rule file's mtime without changing its content.
func (w *pipelineWriter) PokeRuleReload(absPath string) {
	now := time.Now()
	_ = os.Chtimes(absPath, now, now)
}

// DisabledRuleSet returns a snapshot of the currently disabled rule identities.
func (w *pipelineWriter) DisabledRuleSet() map[string]bool {
	return w.identitySet(func(t tenantEntryYAML) []string { return t.DisabledRules })
}

// DisabledPipelineSet returns a snapshot of the currently disabled pipeline identities.
func (w *pipelineWriter) DisabledPipelineSet() map[string]bool {
	return w.identitySet(func(t tenantEntryYAML) []string { return t.DisabledPipelines })
}

func (w *pipelineWriter) identitySet(pick func(tenantEntryYAML) []string) map[string]bool {
	out := make(map[string]bool)
	t, err := w.readTenantEntry()
	if err != nil {
		return out
	}
	for _, id := range pick(t) {
		out[id] = true
	}
	return out
}

// toggleIdentity adds or removes identity from set, keeping it de-duplicated.
func toggleIdentity(set []string, identity string, add bool) []string {
	idx := -1
	for i, v := range set {
		if v == identity {
			idx = i
			break
		}
	}
	if add {
		if idx >= 0 {
			return set
		}
		return append(set, identity)
	}
	if idx < 0 {
		return set
	}
	return append(set[:idx], set[idx+1:]...)
}

func (w *pipelineWriter) tenantLockPath() string {
	return filepath.Join(w.dir, TenantFileName)
}

// ReadPatterns reads patterns.yaml as it is on disk right now, tolerating a
// missing or empty file. No lock is taken: WritePatterns publishes through an
// atomic rename, so a reader observes either the previous file or the complete
// new one, never a partial write.
func (w *pipelineWriter) ReadPatterns() (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(w.dir, PatternsFileName))
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var content patternsFileYAML
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, err
	}
	if content.Patterns == nil {
		return map[string]string{}, nil
	}
	return content.Patterns, nil
}

// WritePatterns atomically rewrites patterns.yaml with the full pattern map.
func (w *pipelineWriter) WritePatterns(patterns map[string]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if patterns == nil {
		patterns = map[string]string{}
	}
	return withFileLock(filepath.Join(w.dir, PatternsFileName), func() error {
		return w.writeYAML(PatternsFileName, patternsFileYAML{Patterns: patterns})
	})
}

func (w *pipelineWriter) writeYAML(name string, v any) error {
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	dst := filepath.Join(w.dir, name)
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

// withFileLock acquires an exclusive advisory lock on <path>.lock via flock,
// runs fn, releases the lock on close. The kernel drops the lock on fd close
// (including process death), so a crashed writer never leaves a stuck lock.
// Linux/BSD only; UTMStack runs in Linux containers. Mirrors the helper in
// modules/integrations/repository/tenant_store.go — duplicated rather than
// shared across those two packages' current boundary.
func withFileLock(path string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lf, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lf.Close()
	if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	return fn()
}
