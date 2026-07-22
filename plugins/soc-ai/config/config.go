package config

import (
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

const (
	processName        = "plugin_com.utmstack.soc-ai"
	pluginFile         = "system_plugins_soc_ai.yaml"
	pipelineDefault    = "/workdir/pipeline"
	instanceConfigPath = "/updates/instance-config.yml"
	reconnectDelay     = 5 * time.Second
	pollInterval       = 30 * time.Second
)

type instanceConfig struct {
	Server      string `yaml:"server"`
	InstanceID  string `yaml:"instance_id"`
	InstanceKey string `yaml:"instance_key"`
}

func readInstanceConfig(path string) (*instanceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ic instanceConfig
	if err := yaml.Unmarshal(data, &ic); err != nil {
		return nil, err
	}
	return &ic, nil
}

type Config struct {
	// Platform (from plugins.PluginCfg("com.utmstack"))
	Backend     string
	InternalKey string

	// LLM (from YAML)
	Provider       string
	URL            string
	Model          string
	APIKey         string // decrypted
	AuthType       string // bearer | header | none
	AuthHeaderName string
	CustomHeaders  map[string]string
	MaxTokens      int

	// Agent behavior (from YAML)
	MaxToolIterations int
	AutoAnalyze       bool
	Capabilities      []string

	// Derived: true when the plugin is configured enough to run.
	ModuleActive bool
}

type fileConfig struct {
	Provider          string            `yaml:"provider"`
	Model             string            `yaml:"model"`
	URL               string            `yaml:"url"`
	APIKey            string            `yaml:"api_key"`
	AuthType          string            `yaml:"auth_type"`
	AuthHeaderName    string            `yaml:"auth_header_name"`
	CustomHeaders     map[string]string `yaml:"custom_headers"`
	MaxTokens         int               `yaml:"max_tokens"`
	MaxToolIterations int               `yaml:"max_tool_iterations"`
	AutoAnalyze       bool              `yaml:"auto_analyze"`
	Capabilities      []string          `yaml:"capabilities"`
}

// pluginsFile matches the flat wrapper the backend writes: plugins.soc-ai.*
type pluginsFile struct {
	Plugins map[string]fileConfig `yaml:"plugins"`
}

const pluginKey = "soc-ai"

var providerDefaultURLs = map[string]string{
	"openai":    "https://api.openai.com/v1/chat/completions",
	"anthropic": "https://api.anthropic.com/v1/messages",
	"azure":     "",
	"gemini":    "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
	"ollama":    "http://localhost:11434/v1/chat/completions",
	"mistral":   "https://api.mistral.ai/v1/chat/completions",
	"deepseek":  "https://api.deepseek.com/chat/completions",
	"groq":      "https://api.groq.com/openai/v1/chat/completions",
}

var sensitiveKeys = map[string]bool{"api_key": true}

var (
	mu      sync.RWMutex
	current Config
	updates = make(chan *Config, 1)
)

func GetConfig() *Config {
	mu.RLock()
	defer mu.RUnlock()
	c := current
	return &c
}

func GetConfigUpdateChannel() <-chan *Config { return updates }

func set(c *Config) {
	mu.Lock()
	current = *c
	mu.Unlock()
	select {
	case updates <- c:
	default:
	}
}

func StartConfigurationSystem() {
	pipelineDir := pipelineDefault
	var encKey, backend, internalKey string

	for {
		cfg := plugins.PluginCfg("com.utmstack")
		if cfg.Exists() {
			if d := cfg.Get("pipelineDir").String(); d != "" {
				pipelineDir = d
			}
			encKey = cfg.Get("encryptionKey").String()
			backend = cfg.Get("backend").String()
			internalKey = cfg.Get("internalKey").String()
			if backend != "" && internalKey != "" {
				break
			}
		}
		_ = catcher.Error("platform configuration not ready (backend/internalKey)", nil, map[string]any{"process": processName})
		time.Sleep(reconnectDelay)
	}

	filePath := filepath.Join(pipelineDir, pluginFile)

	load := func() {
		c := readConfig(filePath, encKey, backend, internalKey)
		set(c)
	}
	load()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_ = catcher.Error("failed to create file watcher; falling back to polling", err, map[string]any{"process": processName})
		for range time.Tick(pollInterval) {
			load()
		}
		return
	}
	defer watcher.Close()

	if err := watcher.Add(pipelineDir); err != nil {
		_ = catcher.Error("failed to watch pipeline dir; falling back to polling", err, map[string]any{"process": processName})
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

func readConfig(path, encKey, backend, internalKey string) *Config {
	c := &Config{Backend: backend, InternalKey: internalKey}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			_ = catcher.Error("failed to read config file", err, map[string]any{"process": processName, "file": path})
		}
		return c // inactive
	}

	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		_ = catcher.Error("failed to parse config file", err, map[string]any{"process": processName, "file": path})
		return c // inactive
	}
	fc, ok := pf.Plugins[pluginKey]
	if !ok {
		return c // inactive: module not configured
	}

	cipher := NewCipher(encKey)
	apiKey := fc.APIKey
	if encKey != "" && sensitiveKeys["api_key"] && apiKey != "" {
		if dec, derr := cipher.Decrypt(apiKey); derr == nil {
			apiKey = dec
		}
	}

	var customHeaders map[string]string
	if len(fc.CustomHeaders) > 0 {
		customHeaders = make(map[string]string, len(fc.CustomHeaders))
		for k, v := range fc.CustomHeaders {
			if encKey != "" {
				if dec, derr := cipher.Decrypt(v); derr == nil {
					v = dec
				}
			}
			customHeaders[k] = v
		}
	}

	url := fc.URL
	if url == "" {
		if def, ok := providerDefaultURLs[fc.Provider]; ok {
			url = def
		}
	}

	maxTokens := fc.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	maxIters := fc.MaxToolIterations
	if maxIters <= 0 {
		maxIters = 12
	}
	authType := fc.AuthType
	if authType == "" {
		authType = "bearer"
	}
	authHeaderName := fc.AuthHeaderName

	if fc.Provider == "threatwinds" {
		// ponytail: hit the CM proxy directly with id+key from /updates/instance-config.yml; user's API key is ignored.
		ic, ierr := readInstanceConfig(instanceConfigPath)
		if ierr != nil || ic.Server == "" || ic.InstanceID == "" || ic.InstanceKey == "" {
			_ = catcher.Error("threatwinds provider requires instance-config.yml", ierr, map[string]any{"process": processName, "file": instanceConfigPath})
			return c // inactive
		}
		url = strings.TrimRight(ic.Server, "/") + "/proxy/api/ai/v1/chat/completions"
		apiKey = ""
		authType = "none"
		authHeaderName = ""
		if customHeaders == nil {
			customHeaders = make(map[string]string, 2)
		}
		customHeaders["id"] = ic.InstanceID
		customHeaders["key"] = ic.InstanceKey
	}

	c.Provider = fc.Provider
	c.URL = url
	c.Model = fc.Model
	c.APIKey = apiKey
	c.AuthType = authType
	c.AuthHeaderName = authHeaderName
	c.CustomHeaders = customHeaders
	c.MaxTokens = maxTokens
	c.MaxToolIterations = maxIters
	c.AutoAnalyze = fc.AutoAnalyze
	c.Capabilities = fc.Capabilities
	c.ModuleActive = fc.Provider != "" && fc.Model != "" && url != ""

	return c
}
