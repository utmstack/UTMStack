package config

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"gopkg.in/yaml.v3"
)

const (
	processName        = "plugin_com.utmstack.feeds"
	pluginFile         = "system_plugins_feeds.yaml"
	pluginKey          = "feeds"
	pipelineDirDefault = "/workdir/pipeline"
	reconnectDelay     = 5 * time.Second
	pollInterval       = 30 * time.Second
)

type PluginConfig struct {
	Enabled   bool
	APIKey    string // decrypted
	APISecret string // decrypted
}

func (c PluginConfig) Configured() bool { return c.APIKey != "" && c.APISecret != "" }

type fileConfig struct {
	Enabled   bool   `yaml:"enabled"`
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
}

// pluginsFile matches the wrapper the backend writes: plugins.feeds.*
type pluginsFile struct {
	Plugins map[string]fileConfig `yaml:"plugins"`
}

var (
	mu      sync.RWMutex
	current PluginConfig
)

func GetPluginConfig() PluginConfig {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

func setPluginConfig(c PluginConfig) {
	mu.Lock()
	current = c
	mu.Unlock()
}

func StartConfigurationSystem() {
	pipelineDir := pipelineDirDefault
	var encKey string

	for {
		cfg := plugins.PluginCfg("com.utmstack")
		if cfg.Exists() {
			if d := cfg.Get("pipelineDir").String(); d != "" {
				pipelineDir = d
			}
			encKey = cfg.Get("encryptionKey").String()
			break
		}
		_ = catcher.Error("platform configuration not ready", nil, map[string]any{"process": processName})
		time.Sleep(reconnectDelay)
	}

	filePath := filepath.Join(pipelineDir, pluginFile)
	load := func() { setPluginConfig(readPluginConfig(filePath, encKey)) }
	load()

	go watch(pipelineDir, filePath, load)
}

func watch(pipelineDir, filePath string, load func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_ = catcher.Error("failed to create file watcher; falling back to polling", err, map[string]any{"process": processName})
		for range time.Tick(pollInterval) {
			load()
		}
		return
	}
	defer func() { _ = watcher.Close() }()

	if err := watcher.Add(pipelineDir); err != nil {
		_ = catcher.Error("failed to watch the pipeline dir; falling back to polling", err, map[string]any{"process": processName})
		for range time.Tick(pollInterval) {
			load()
		}
		return
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name == filePath && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename)) {
				load()
			}
		case <-ticker.C:
			load()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			_ = catcher.Error("file watcher error", err, map[string]any{"process": processName})
		}
	}
}

func readPluginConfig(path, encKey string) PluginConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			_ = catcher.Error("failed to read the configuration file", err,
				map[string]any{"process": processName, "file": path})
		}
		return PluginConfig{}
	}

	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		_ = catcher.Error("failed to parse the configuration file", err,
			map[string]any{"process": processName, "file": path})
		return PluginConfig{}
	}

	fc, ok := pf.Plugins[pluginKey]
	if !ok {
		return PluginConfig{}
	}

	c := PluginConfig{Enabled: fc.Enabled, APIKey: fc.APIKey, APISecret: fc.APISecret}
	if encKey == "" {
		return c
	}

	cipher := NewCipher(encKey)
	if dec, err := cipher.Decrypt(fc.APIKey); err == nil {
		c.APIKey = dec
	}
	if dec, err := cipher.Decrypt(fc.APISecret); err == nil {
		c.APISecret = dec
	}
	return c
}
