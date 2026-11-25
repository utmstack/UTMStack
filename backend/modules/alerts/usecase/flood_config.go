package usecase

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	floodConfigFile = "system_plugins_rule-flood-guard.yaml"
	floodConfigKey  = "rule-flood-guard"

	defaultFloodEnabled         = true
	defaultFloodThreshold int64 = 50
	defaultFloodWindowH         = 24
	defaultFloodIntervalS       = 300
)

// FloodConfig is what the flood guard reads on every pass, so a change to the file takes effect on
// the next one without restarting anything.
type FloodConfig struct {
	Enabled     bool
	Threshold   int64
	WindowHours int
	Interval    time.Duration
}

func (c FloodConfig) window() time.Duration { return time.Duration(c.WindowHours) * time.Hour }

type floodFileConfig struct {
	Enabled         *bool  `yaml:"enabled"`
	Threshold       *int64 `yaml:"threshold"`
	WindowHours     *int   `yaml:"windowHours"`
	IntervalSeconds *int   `yaml:"intervalSeconds"`
}

type floodPluginsFile struct {
	Plugins map[string]floodFileConfig `yaml:"plugins"`
}

// DefaultFloodConfig is what applies when the file does not exist or leaves a value out.
func DefaultFloodConfig() FloodConfig {
	return FloodConfig{
		Enabled:     defaultFloodEnabled,
		Threshold:   defaultFloodThreshold,
		WindowHours: defaultFloodWindowH,
		Interval:    defaultFloodIntervalS * time.Second,
	}
}

// LoadFloodConfig reads the settings from the pipeline folder. The file is the one the plugin this
// replaces read, so an installation keeps the values it already has. A missing file means defaults;
// a value that is not positive is ignored.
func LoadFloodConfig(dir string) (FloodConfig, error) {
	cfg := DefaultFloodConfig()

	data, err := os.ReadFile(filepath.Join(dir, floodConfigFile))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}

		return cfg, err
	}

	var file floodPluginsFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return DefaultFloodConfig(), err
	}

	fc, ok := file.Plugins[floodConfigKey]
	if !ok {
		return cfg, nil
	}

	if fc.Enabled != nil {
		cfg.Enabled = *fc.Enabled
	}

	if fc.Threshold != nil && *fc.Threshold > 0 {
		cfg.Threshold = *fc.Threshold
	}

	if fc.WindowHours != nil && *fc.WindowHours > 0 {
		cfg.WindowHours = *fc.WindowHours
	}

	if fc.IntervalSeconds != nil && *fc.IntervalSeconds > 0 {
		cfg.Interval = time.Duration(*fc.IntervalSeconds) * time.Second
	}

	return cfg, nil
}
