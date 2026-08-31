package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/shared/crypto"
	"gopkg.in/yaml.v3"
)

const (
	pluginFile         = "system_plugins_sophos.yaml"
	pluginName         = "com.utmstack.sophos"
	processName        = "plugin_" + pluginName
	pipelineDirDefault = "/workdir/pipeline"
)

type ConfigurationSection struct {
	ModuleActive bool
	ModuleGroups []*ModuleGroup
}

type ModuleGroup struct {
	Id                        int32
	GroupName                 string
	TenantId                  string
	ModuleGroupConfigurations []*Configuration
}

func (g *ModuleGroup) Key() string { return g.TenantId + "/" + g.GroupName }

type Configuration struct {
	ConfKey   string
	ConfValue string
}

var (
	cnf              *ConfigurationSection
	mu               sync.Mutex
	configUpdateChan chan *ConfigurationSection

	encKeyMu  sync.RWMutex
	encKeyVal string
)

// setEncryptionKey records the plugin-wide encryption key. Cursor payloads are
// encrypted with it because the ingest_cursors bucket has no auth and no TLS.
func setEncryptionKey(key string) {
	encKeyMu.Lock()
	defer encKeyMu.Unlock()
	encKeyVal = key
}

func getEncryptionKey() string {
	encKeyMu.RLock()
	defer encKeyMu.RUnlock()
	return encKeyVal
}

func init() {
	configUpdateChan = make(chan *ConfigurationSection, 1)
}

// GetConfig returns the current configuration (nil-safe).
func GetConfig() *ConfigurationSection {
	mu.Lock()
	defer mu.Unlock()
	if cnf == nil {
		return &ConfigurationSection{}
	}
	return cnf
}

// GetConfigUpdateChannel returns the channel that receives config updates.
func GetConfigUpdateChannel() <-chan *ConfigurationSection {
	return configUpdateChan
}

// StartConfigurationSystem blocks until configured, then watches indefinitely.
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
			setEncryptionKey(encKey)
			break
		}
		_ = catcher.Error("plugin configuration not found", nil, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
	}

	filePath := filepath.Join(pipelineDir, pluginFile)

	if sec := readConfig(filePath, encKey); sec != nil {
		mu.Lock()
		cnf = sec
		mu.Unlock()
		push(sec)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_ = catcher.Error("failed to create file watcher", err, map[string]any{"process": processName})
		pollFallback(filePath, encKey)
		return
	}
	defer watcher.Close()

	// Watch the directory, not the file, to catch atomic write (rename) events.
	if err := watcher.Add(pipelineDir); err != nil {
		_ = catcher.Error("failed to watch pipeline dir", err, map[string]any{"process": processName})
		pollFallback(filePath, encKey)
		return
	}

	catcher.Info(fmt.Sprintf("watching %s for config changes", filePath), map[string]any{"process": processName})

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name != filePath {
				continue
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) ||
				event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
				if sec := readConfig(filePath, encKey); sec != nil {
					mu.Lock()
					cnf = sec
					mu.Unlock()
					push(sec)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			_ = catcher.Error("watcher error", err, map[string]any{"process": processName})
		}
	}
}

func pollFallback(filePath, encKey string) {
	catcher.Warn("falling back to 30s polling", map[string]any{"process": processName})
	for range time.Tick(30 * time.Second) {
		if sec := readConfig(filePath, encKey); sec != nil {
			mu.Lock()
			cnf = sec
			mu.Unlock()
			push(sec)
		}
	}
}

func push(sec *ConfigurationSection) {
	select {
	case configUpdateChan <- sec:
	default:
	}
}

type configGroupYAML struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Config      map[string]string `yaml:"config"`
}

type tenantYAML struct {
	ID     string            `yaml:"id"`
	Groups []configGroupYAML `yaml:"groups"`
}

type pluginsFile struct {
	Plugins map[string]struct {
		Tenants []tenantYAML `yaml:"tenants"`
	} `yaml:"plugins"`
}

const pluginKey = "sophos"

var sensitiveKeys = map[string]bool{
	"sophos_x_api_key": true,
}

func readConfig(path, encKey string) *ConfigurationSection {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// A removed file means the module is disabled, not a read failure.
			return &ConfigurationSection{ModuleActive: false}
		}
		_ = catcher.Error("failed to read config file", err, map[string]any{"process": processName, "file": path})
		return nil
	}

	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		_ = catcher.Error("failed to parse config file", err, map[string]any{"process": processName, "file": path})
		return nil
	}
	tenants := pf.Plugins[pluginKey].Tenants

	sec := &ConfigurationSection{}
	var id int32
	for _, tc := range tenants {
		for _, g := range tc.Groups {
			id++
			grp := &ModuleGroup{
				Id:        id,
				GroupName: g.Name,
				TenantId:  tc.ID,
			}
			for k, v := range g.Config {
				conf := &Configuration{ConfKey: k, ConfValue: v}
				if encKey != "" && sensitiveKeys[conf.ConfKey] {
					if dec, err := crypto.NewCipher(encKey).Decrypt(conf.ConfValue); err == nil {
						conf.ConfValue = dec
					}
				}
				grp.ModuleGroupConfigurations = append(grp.ModuleGroupConfigurations, conf)
			}
			sec.ModuleGroups = append(sec.ModuleGroups, grp)
		}
	}
	sec.ModuleActive = len(sec.ModuleGroups) > 0
	return sec
}
