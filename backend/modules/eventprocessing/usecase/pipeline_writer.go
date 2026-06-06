package usecase

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// pipelineWriter is the single writer for tenants.yaml and patterns.yaml in the
// pipeline directory shared with the event processor. The engine reads these via
// loadCfg (Walk + ReadPbYaml → protojson.Unmarshal), so the YAML field names
// must be the protobuf JSON names (camelCase as defined in plugins.proto).
type pipelineWriter struct {
	dir string
	mu  sync.Mutex
}

func NewPipelineWriter(dir string) *pipelineWriter {
	return &pipelineWriter{dir: dir}
}

// --- YAML structs (proto JSON field names) ---

type tenantFileYAML struct {
	Tenants []tenantEntryYAML `yaml:"tenants"`
}

type tenantEntryYAML struct {
	ID     string          `yaml:"id"`
	Name   string          `yaml:"name"`
	Assets []assetFileYAML `yaml:"assets,omitempty"`
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

// WriteTenants atomically rewrites tenants.yaml with the full asset list.
// There is always exactly one tenant ("Default"); assets are its CIA-config entries.
func (w *pipelineWriter) WriteTenants(assets []assetFileYAML) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	content := tenantFileYAML{
		Tenants: []tenantEntryYAML{{
			ID:     DefaultTenantID,
			Name:   DefaultTenantName,
			Assets: assets,
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
