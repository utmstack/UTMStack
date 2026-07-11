package usecase

import (
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

// pipelineWriter is the single writer for tenants.yaml and patterns.yaml in the
// pipeline directory shared with the event processor. The engine reads these via
// loadCfg (Walk + ReadPbYaml → protojson.Unmarshal), so the YAML field names
// must be the protobuf JSON names (camelCase as defined in plugins.proto).
//
// tenants.yaml has two independent writers (asset CIA config and rule
// enable/disable state), so the last-written value of each is kept in memory
// and combined on every write — otherwise whichever writes last would wipe
// out the other's data. Both are seeded from the file on disk at construction
// so a fresh process picks up whatever a previous run persisted.
type pipelineWriter struct {
	dir string
	mu  sync.Mutex

	assets        []assetFileYAML
	disabledRules map[string]bool // rule identity (basename, no ext) -> disabled
}

func NewPipelineWriter(dir string) *pipelineWriter {
	w := &pipelineWriter{dir: dir, disabledRules: make(map[string]bool)}
	w.seedFromDisk()
	return w
}

// seedFromDisk loads the current tenants.yaml (if any) into memory so the
// first WriteTenants/SetRuleDisabled call after a restart doesn't clobber
// whatever the other writer had persisted in a previous run.
func (w *pipelineWriter) seedFromDisk() {
	data, err := os.ReadFile(filepath.Join(w.dir, TenantFileName))
	if err != nil {
		return
	}
	var content tenantFileYAML
	if err := yaml.Unmarshal(data, &content); err != nil || len(content.Tenants) == 0 {
		return
	}
	t := content.Tenants[0]
	w.assets = t.Assets
	for _, id := range t.DisabledRules {
		w.disabledRules[id] = true
	}
}

// --- YAML structs (proto JSON field names) ---

type tenantFileYAML struct {
	Tenants []tenantEntryYAML `yaml:"tenants"`
}

type tenantEntryYAML struct {
	ID            string          `yaml:"id"`
	Name          string          `yaml:"name"`
	Assets        []assetFileYAML `yaml:"assets,omitempty"`
	DisabledRules []string        `yaml:"disabledRules,omitempty"`
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

// WriteTenants atomically rewrites tenants.yaml with the full asset list,
// keeping whatever disabled-rules state is already known.
// There is always exactly one tenant ("Default"); assets are its CIA-config entries.
func (w *pipelineWriter) WriteTenants(assets []assetFileYAML) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.assets = assets
	return w.writeTenantsLocked()
}

// SetRuleDisabled records a rule (identified the same way the event processor
// does: filename without directory or extension) as disabled/enabled and
// persists it to tenants.yaml, keeping the current asset list untouched.
func (w *pipelineWriter) SetRuleDisabled(identity string, disabled bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if disabled {
		w.disabledRules[identity] = true
	} else {
		delete(w.disabledRules, identity)
	}
	return w.writeTenantsLocked()
}

// DisabledRuleSet returns a snapshot of the currently disabled rule identities.
func (w *pipelineWriter) DisabledRuleSet() map[string]bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make(map[string]bool, len(w.disabledRules))
	for k := range w.disabledRules {
		out[k] = true
	}
	return out
}

func (w *pipelineWriter) writeTenantsLocked() error {
	names := make([]string, 0, len(w.disabledRules))
	for id := range w.disabledRules {
		names = append(names, id)
	}
	sort.Strings(names)

	content := tenantFileYAML{
		Tenants: []tenantEntryYAML{{
			ID:            DefaultTenantID,
			Name:          DefaultTenantName,
			Assets:        w.assets,
			DisabledRules: names,
		}},
	}
	return w.writeYAML(TenantFileName, content)
}

// WritePatterns atomically rewrites patterns.yaml with the full pattern map.
func (w *pipelineWriter) WritePatterns(patterns map[string]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if patterns == nil {
		patterns = map[string]string{}
	}
	return w.writeYAML(PatternsFileName, patternsFileYAML{Patterns: patterns})
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
