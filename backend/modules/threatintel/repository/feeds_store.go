package repository

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	configFileName = "system_plugins_feeds.yaml"
	pluginKey      = "feeds"
)

// FeedsConfig is what the feeds plugin needs to know: whether this instance
// sends its incidents to ThreatWinds, and the credentials it registered with.
// The two secrets are stored encrypted, and the plugin decrypts them with the
// same instance key.
type FeedsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	APIKey    string `yaml:"api_key,omitempty"`
	APISecret string `yaml:"api_secret,omitempty"`
}

// pluginsFile is the on-disk wrapper every plugin config shares —
// plugins.<key>.* — so this file looks like the ones beside it.
type pluginsFile struct {
	Plugins map[string]FeedsConfig `yaml:"plugins"`
}

// ConfigStore owns system_plugins_feeds.yaml. The plugin only reads it, which
// is the point: credentials never travel back over the API, where a secret is
// withheld by design.
type ConfigStore struct {
	dir string
	mu  sync.Mutex
}

func NewConfigStore(dir string) *ConfigStore { return &ConfigStore{dir: dir} }

func (s *ConfigStore) path() string { return filepath.Join(s.dir, configFileName) }

func (s *ConfigStore) Load() (FeedsConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.read()
	if err != nil || cfg == nil {
		return FeedsConfig{}, err
	}
	return *cfg, nil
}

// Update applies a change to whatever is on disk. Enabling the plugin and
// storing credentials are separate operations on the same file, so each one
// re-reads under the lock instead of writing a whole config it did not see.
func (s *ConfigStore) Update(apply func(*FeedsConfig)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.withFileLock(func() error {
		cfg, err := s.read()
		if err != nil {
			return err
		}
		if cfg == nil {
			cfg = &FeedsConfig{}
		}
		apply(cfg)
		return s.write(cfg)
	})
}

func (s *ConfigStore) read() (*FeedsConfig, error) {
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

func (s *ConfigStore) write(cfg *FeedsConfig) error {
	data, err := yaml.Marshal(pluginsFile{Plugins: map[string]FeedsConfig{pluginKey: *cfg}})
	if err != nil {
		return err
	}

	// A private temp file that is renamed into place: a reader watching this
	// path sees either the old file or the new one, never half of either.
	tmp, err := os.CreateTemp(s.dir, configFileName+".*.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path())
}
