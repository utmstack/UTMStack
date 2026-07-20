package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Enabled         bool
	Threshold       int64
	WindowHours     int
	IntervalSeconds int
	BackendURL      string
	InternalKey     string
}

const (
	pluginKey          = "rule-flood-guard"
	configFileName     = "system_plugins_rule-flood-guard.yaml"
	pipelineDirDefault = "/workdir/pipeline"
	processName        = "plugin_com.utmstack.rule-flood-guard"

	defaultEnabled               = true
	defaultThreshold       int64 = 50
	defaultWindowHours           = 24
	defaultIntervalSeconds       = 300
)

type fileConfig struct {
	Enabled         *bool  `yaml:"enabled"`
	Threshold       *int64 `yaml:"threshold"`
	WindowHours     *int   `yaml:"windowHours"`
	IntervalSeconds *int   `yaml:"intervalSeconds"`
}

type pluginsFile struct {
	Plugins map[string]fileConfig `yaml:"plugins"`
}

func defaultKnobs() Config {
	return Config{
		Enabled:         defaultEnabled,
		Threshold:       defaultThreshold,
		WindowHours:     defaultWindowHours,
		IntervalSeconds: defaultIntervalSeconds,
	}
}

func applyFileConfig(base Config, fc fileConfig) Config {
	if fc.Enabled != nil {
		base.Enabled = *fc.Enabled
	}
	if fc.Threshold != nil && *fc.Threshold > 0 {
		base.Threshold = *fc.Threshold
	}
	if fc.WindowHours != nil && *fc.WindowHours > 0 {
		base.WindowHours = *fc.WindowHours
	}
	if fc.IntervalSeconds != nil && *fc.IntervalSeconds > 0 {
		base.IntervalSeconds = *fc.IntervalSeconds
	}
	return base
}

func loadKnobsFromFile(path string) (Config, error) {
	knobs := defaultKnobs()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return knobs, nil
		}
		return knobs, err
	}

	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return knobs, err
	}
	fc, ok := pf.Plugins[pluginKey]
	if !ok {
		return knobs, nil
	}
	return applyFileConfig(knobs, fc), nil
}

type configHolder struct {
	mu  sync.RWMutex
	cfg Config
}

func newConfigHolder(initial Config) *configHolder {
	return &configHolder{cfg: initial}
}

func (h *configHolder) Get() Config {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.cfg
}

func (h *configHolder) setKnobs(knobs Config) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cfg.Enabled = knobs.Enabled
	h.cfg.Threshold = knobs.Threshold
	h.cfg.WindowHours = knobs.WindowHours
	h.cfg.IntervalSeconds = knobs.IntervalSeconds
}

func loadConfig() (Config, string) {
	cfg := plugins.PluginCfg("com.utmstack")
	backendURL := normalizeBackendURL(cfg.Get("backend").String())
	internalKey := cfg.Get("internalKey").String()
	pipelineDir := cfg.Get("pipelineDir").String()
	if pipelineDir == "" {
		pipelineDir = pipelineDirDefault
	}

	path := filepath.Join(pipelineDir, configFileName)
	knobs, err := loadKnobsFromFile(path)
	if err != nil {
		_ = catcher.Error("rule-flood-guard: failed to read config file, using defaults", err, map[string]any{"process": processName, "file": path})
		knobs = defaultKnobs()
	}
	knobs.BackendURL = backendURL
	knobs.InternalKey = internalKey
	return knobs, path
}

func normalizeBackendURL(raw string) string {
	if raw == "" {
		return raw
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return "http://" + raw
}

func watchConfigFile(ctx context.Context, holder *configHolder, path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_ = catcher.Error("rule-flood-guard: failed to create config file watcher, hot-reload disabled", err, map[string]any{"process": processName})
		return
	}
	defer func() { _ = watcher.Close() }()

	dir := filepath.Dir(path)
	if err := watcher.Add(dir); err != nil {
		_ = catcher.Error("rule-flood-guard: failed to watch config dir, hot-reload disabled", err, map[string]any{"process": processName, "dir": dir})
		return
	}

	catcher.Info("rule-flood-guard: watching "+path+" for config changes", map[string]any{"process": processName})

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name != path {
				continue
			}
			if !(event.Has(fsnotify.Write) || event.Has(fsnotify.Create) ||
				event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove)) {
				continue
			}
			knobs, err := loadKnobsFromFile(path)
			if err != nil {
				_ = catcher.Error("rule-flood-guard: failed to reload config file, keeping previous values", err, map[string]any{"process": processName, "file": path})
				continue
			}
			holder.setKnobs(knobs)
			catcher.Info("rule-flood-guard: config reloaded", map[string]any{
				"process": processName, "enabled": knobs.Enabled, "threshold": knobs.Threshold,
				"windowHours": knobs.WindowHours, "intervalSeconds": knobs.IntervalSeconds,
			})
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			_ = catcher.Error("rule-flood-guard: config watcher error", err, map[string]any{"process": processName})
		case <-ctx.Done():
			return
		}
	}
}

func (c Config) tickInterval() time.Duration {
	return time.Duration(c.IntervalSeconds) * time.Second
}

func (c Config) window() time.Duration {
	return time.Duration(c.WindowHours) * time.Hour
}
