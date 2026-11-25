package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/integrations/shared/crypto"
	"github.com/utmstack/UTMStack/integrations/shared/modulecfg"
	"gopkg.in/yaml.v3"
)

const (
	pluginName  = "com.utmstack.o365_graph"
	processName = "plugin_" + pluginName
	pluginKey   = "o365_graph"
	// DefaultTenant mirrors o365's: it is the tenant id enqueued logs get when a
	// group is configured without one.
	DefaultTenant = "ce66672c-e36d-4761-a8c8-90058fee1a24"
)

type CloudEnvironment string

const (
	CloudCommercial CloudEnvironment = "Commercial"
	CloudGCC        CloudEnvironment = "GCC"
	CloudGCCHigh    CloudEnvironment = "GCCHigh"
	CloudDoD        CloudEnvironment = "DoD"
)

type CloudConfig struct {
	LoginAuthority string
	GraphEndpoint  string
	GraphScope     string
}

func GetCloudConfig(env CloudEnvironment) CloudConfig {
	configs := map[CloudEnvironment]CloudConfig{
		CloudCommercial: {
			LoginAuthority: "https://login.microsoftonline.com/",
			GraphEndpoint:  "https://graph.microsoft.com",
			GraphScope:     "https://graph.microsoft.com/.default",
		},
		// GCC tenants are registered in a worldwide Azure AD directory, so the
		// login authority and Graph endpoint stay global.
		CloudGCC: {
			LoginAuthority: "https://login.microsoftonline.com/",
			GraphEndpoint:  "https://graph.microsoft.com",
			GraphScope:     "https://graph.microsoft.com/.default",
		},
		CloudGCCHigh: {
			LoginAuthority: "https://login.microsoftonline.us/",
			GraphEndpoint:  "https://graph.microsoft.us",
			GraphScope:     "https://graph.microsoft.us/.default",
		},
		CloudDoD: {
			LoginAuthority: "https://login.microsoftonline.us/",
			GraphEndpoint:  "https://dod-graph.microsoft.us",
			GraphScope:     "https://dod-graph.microsoft.us/.default",
		},
	}

	config, exists := configs[env]
	if !exists {
		return configs[CloudCommercial]
	}
	return config
}

type ConfigurationSection struct {
	ModuleActive bool
	ModuleGroups []*ModuleGroup
}

type ModuleGroup struct {
	Id                        int32
	GroupName                 string
	UtmTenantId               string
	ModuleGroupConfigurations []*Configuration
}

func (g *ModuleGroup) Key() string { return g.UtmTenantId + "/" + g.GroupName }

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

// The same key decrypts sensitive config values and encrypts the queue-path
// cursor payload before it crosses NATS.
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

func GetConfig() *ConfigurationSection {
	mu.Lock()
	defer mu.Unlock()
	if cnf == nil {
		return &ConfigurationSection{}
	}
	return cnf
}

func GetConfigUpdateChannel() <-chan *ConfigurationSection {
	return configUpdateChan
}

func StartConfigurationSystem() {
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		_ = catcher.Error("ENCRYPTION_KEY is not set; the module configuration cannot be read", nil, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	setEncryptionKey(encKey)

	err := modulecfg.Watch(context.Background(), os.Getenv("NATS_URL"), "o365_graph", crypto.NewCipher(encKey),
		func(data []byte) {
			sec := parseConfig(data, encKey)
			if sec == nil {
				return
			}

			mu.Lock()
			cnf = sec
			mu.Unlock()
			push(sec)
		},
		func(err error) {
			_ = catcher.Error("module configuration could not be read", err, map[string]any{"process": processName})
		})
	if err != nil {
		_ = catcher.Error("module configuration watch failed", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
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

var sensitiveKeys = map[string]bool{
	"o365_graph_client_secret": true,
}

func parseConfig(data []byte, encKey string) *ConfigurationSection {
	if data == nil {
		// No stored configuration means the module is disabled, not an error.
		return &ConfigurationSection{ModuleActive: false}
	}

	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		_ = catcher.Error("failed to parse config file", err, map[string]any{"process": processName})
		return nil
	}
	tenants := pf.Plugins[pluginKey].Tenants

	sec := &ConfigurationSection{}
	var id int32
	for _, tc := range tenants {
		for _, g := range tc.Groups {
			id++
			grp := &ModuleGroup{
				Id:          id,
				GroupName:   g.Name,
				UtmTenantId: tc.ID,
			}
			for k, v := range g.Config {
				conf := &Configuration{ConfKey: k, ConfValue: v}
				if sensitiveKeys[k] && encKey != "" {
					dec, err := crypto.NewCipher(encKey).Decrypt(conf.ConfValue)
					if err == nil {
						conf.ConfValue = dec
					}
				}
				grp.ModuleGroupConfigurations = append(grp.ModuleGroupConfigurations, conf)
			}
			sec.ModuleGroups = append(sec.ModuleGroups, grp)
		}
	}
	// A tenant section with no groups must not read as configured.
	sec.ModuleActive = len(sec.ModuleGroups) > 0
	return sec
}
