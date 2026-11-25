package repository

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	configFileName = "system_plugins_soc_ai.yaml"
	pluginKey      = "soc-ai"
)

type FileConfig struct {
	Provider          string            `yaml:"provider"`
	Model             string            `yaml:"model"`
	URL               string            `yaml:"url"`
	APIKey            string            `yaml:"api_key"`
	AuthType          string            `yaml:"auth_type"`
	AuthHeaderName    string            `yaml:"auth_header_name"`
	CustomHeaders     map[string]string `yaml:"custom_headers,omitempty"`
	MaxTokens         int               `yaml:"max_tokens"`
	MaxToolIterations int               `yaml:"max_tool_iterations"`
	AutoAnalyze       bool              `yaml:"auto_analyze"`
	AllowIncidents    bool              `yaml:"allow_incidents"`
}

type socAIConfig struct {
	FileConfig `yaml:",inline"`
	Tenants    map[string]FileConfig `yaml:"tenants,omitempty"`
}

// pluginsFile is the on-disk wrapper shared with the other system_plugins_*.yaml
// files: plugins.<key>.* — keeps soc-ai's config file the same namespaced shape
// as the datasource plugins instead of being the one bare/flat outlier.
type pluginsFile struct {
	Plugins map[string]socAIConfig `yaml:"plugins"`
}

type ConfigStore struct {
	dir string
	mu  sync.Mutex

	onChange func()
}

// SetOnChange registers what to do after the configuration is written. It runs outside the store's lock
// and must not block.
func (s *ConfigStore) SetOnChange(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onChange = f
}

func (s *ConfigStore) changed() {
	s.mu.Lock()
	f := s.onChange
	s.mu.Unlock()

	if f != nil {
		f()
	}
}

func NewConfigStore(dir string) *ConfigStore {
	return &ConfigStore{dir: dir}
}

func (s *ConfigStore) path() string {
	return filepath.Join(s.dir, configFileName)
}

func (s *ConfigStore) Load(tenantID string) (*FileConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil || cfg == nil {
		return nil, err
	}

	if tenantID != "" {
		if own, ok := cfg.Tenants[tenantID]; ok {
			return &own, nil
		}
	}

	return &cfg.FileConfig, nil
}

func (s *ConfigStore) LoadOwn(tenantID string) (*FileConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil || cfg == nil {
		return nil, err
	}
	if tenantID == "" {
		return &cfg.FileConfig, nil
	}
	if own, ok := cfg.Tenants[tenantID]; ok {
		return &own, nil
	}
	return nil, nil
}

func (s *ConfigStore) Save(tenantID string, fc *FileConfig) error {
	err := s.save(tenantID, fc)
	if err == nil {
		s.changed()
	}

	return err
}

func (s *ConfigStore) save(tenantID string, fc *FileConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.withFileLock(func() error {
		cfg, err := s.read()
		if err != nil {
			return err
		}
		if cfg == nil {
			cfg = &socAIConfig{}
		}

		if tenantID == "" {
			cfg.FileConfig = *fc
		} else {
			if cfg.Tenants == nil {
				cfg.Tenants = map[string]FileConfig{}
			}
			cfg.Tenants[tenantID] = *fc
		}

		return s.write(cfg)
	})
}

var ErrNoInstanceDefault = errors.New("socai: the instance default cannot be removed")

func (s *ConfigStore) Delete(tenantID string) error {
	if tenantID == "" {
		return ErrNoInstanceDefault
	}

	err := s.deleteTenant(tenantID)
	if err == nil {
		s.changed()
	}

	return err
}

func (s *ConfigStore) deleteTenant(tenantID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.withFileLock(func() error {
		cfg, err := s.read()
		if err != nil || cfg == nil {
			return err
		}
		if _, ok := cfg.Tenants[tenantID]; !ok {
			return nil
		}

		delete(cfg.Tenants, tenantID)

		return s.write(cfg)
	})
}

func (s *ConfigStore) read() (*socAIConfig, error) {
	data, err := os.ReadFile(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, err
	}
	cfg, ok := pf.Plugins[pluginKey]
	if !ok {
		return nil, nil
	}
	return &cfg, nil
}

func (s *ConfigStore) write(cfg *socAIConfig) error {
	data, err := yaml.Marshal(pluginsFile{Plugins: map[string]socAIConfig{pluginKey: *cfg}})
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(s.dir, configFileName+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), s.path())
}

// Snapshot is the whole configuration as the SOC AI service reads it: the instance default and the
// configuration of each tenant that has its own. Secrets are still encrypted as they are stored.
type Snapshot struct {
	Default *SnapshotConfig           `json:"default,omitempty"`
	Tenants map[string]SnapshotConfig `json:"tenants,omitempty"`
}

// SnapshotConfig is one configuration, with the names the service reads.
type SnapshotConfig struct {
	Provider          string            `json:"provider"`
	Model             string            `json:"model"`
	URL               string            `json:"url"`
	APIKey            string            `json:"api_key,omitempty"`
	AuthType          string            `json:"auth_type,omitempty"`
	AuthHeaderName    string            `json:"auth_header_name,omitempty"`
	CustomHeaders     map[string]string `json:"custom_headers,omitempty"`
	MaxTokens         int               `json:"max_tokens,omitempty"`
	MaxToolIterations int               `json:"max_tool_iterations,omitempty"`
	AutoAnalyze       bool              `json:"auto_analyze"`
	AllowIncidents    bool              `json:"allow_incidents"`
}

func snapshotOf(fc FileConfig) SnapshotConfig {
	return SnapshotConfig{
		Provider: fc.Provider, Model: fc.Model, URL: fc.URL, APIKey: fc.APIKey, AuthType: fc.AuthType,
		AuthHeaderName: fc.AuthHeaderName, CustomHeaders: fc.CustomHeaders, MaxTokens: fc.MaxTokens,
		MaxToolIterations: fc.MaxToolIterations, AutoAnalyze: fc.AutoAnalyze, AllowIncidents: fc.AllowIncidents,
	}
}

// Snapshot reads the configuration for publishing. It is nil when nothing is configured at all.
func (s *ConfigStore) Snapshot() (*Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil || cfg == nil {
		return nil, err
	}

	out := &Snapshot{}

	if cfg.Provider != "" || cfg.Model != "" || cfg.URL != "" {
		d := snapshotOf(cfg.FileConfig)
		out.Default = &d
	}

	if len(cfg.Tenants) > 0 {
		out.Tenants = make(map[string]SnapshotConfig, len(cfg.Tenants))

		for id, fc := range cfg.Tenants {
			out.Tenants[id] = snapshotOf(fc)
		}
	}

	return out, nil
}
