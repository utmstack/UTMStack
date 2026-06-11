package repository

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

const configFileName = "system_plugins_soc_ai.yaml"

type FileConfig struct {
	Provider                  string            `yaml:"provider"`
	Model                     string            `yaml:"model"`
	URL                       string            `yaml:"url"`
	APIKey                    string            `yaml:"api_key"`
	AuthType                  string            `yaml:"auth_type"`
	AuthHeaderName            string            `yaml:"auth_header_name"`
	CustomHeaders             map[string]string `yaml:"custom_headers,omitempty"`
	MaxTokens                 int               `yaml:"max_tokens"`
	MaxToolIterations         int               `yaml:"max_tool_iterations"`
	AutoAnalyze               bool              `yaml:"auto_analyze"`
	ChangeAlertStatus         bool              `yaml:"change_alert_status"`
	AutomaticIncidentCreation bool              `yaml:"automatic_incident_creation"`
	AllowedWriteTools         []string          `yaml:"allowed_write_tools,omitempty"`
}

type ConfigStore struct {
	dir string
	mu  sync.Mutex
}

func NewConfigStore(dir string) *ConfigStore {
	return &ConfigStore{dir: dir}
}

func (s *ConfigStore) path() string {
	return filepath.Join(s.dir, configFileName)
}

func (s *ConfigStore) Load() (*FileConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var fc FileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return nil, err
	}
	return &fc, nil
}

func (s *ConfigStore) Save(fc *FileConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(fc)
	if err != nil {
		return err
	}
	dst := s.path()
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
