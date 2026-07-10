package internal

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"
)

type InstanceConfig struct {
	Server      string `yaml:"server"`
	InstanceID  string `yaml:"instance_id"`
	InstanceKey string `yaml:"instance_key"`
}

var (
	instanceConfig *InstanceConfig
	configOnce     sync.Once
)

// LoadInstanceConfig reads instance-config.yml from updatesDir and caches it.
// Returns an error if the file is missing or cannot be parsed.
// Returns nil and valid config only if all fields are present.
func LoadInstanceConfig(updatesDir string) (*InstanceConfig, error) {
	var err error
	configOnce.Do(func() {
		data, err := os.ReadFile(filepath.Join(updatesDir, "instance-config.yml"))
		if err != nil {
			_ = catcher.Error("error loading instance configuration", err, map[string]any{})
			return
		}
		var cfg InstanceConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			_ = catcher.Error("error loading instance configuration", err, map[string]any{})
			return
		}
		instanceConfig = &cfg
	})
	return instanceConfig, err
}

// Get returns the cached instance config, or nil if not loaded.
func Get() *InstanceConfig {
	return instanceConfig
}
