package instanceconfig

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type InstanceConfig struct {
	Server      string `yaml:"server"`
	InstanceID  string `yaml:"instance_id"`
	InstanceKey string `yaml:"instance_key"`
}

var (
	mu         sync.RWMutex
	updatesDir string
	cached     *InstanceConfig
)

// Init records where instance-config.yml lives. Call once at startup.
func Init(dir string) {
	mu.Lock()
	updatesDir = dir
	mu.Unlock()
}

func Get() *InstanceConfig {
	mu.RLock()
	if cached != nil {
		defer mu.RUnlock()
		return cached
	}
	dir := updatesDir
	mu.RUnlock()

	if dir == "" {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(dir, "instance-config.yml"))
	if err != nil {
		return nil
	}
	var cfg InstanceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	mu.Lock()
	cached = &cfg
	mu.Unlock()
	return &cfg
}
