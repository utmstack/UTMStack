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
	pluginName  = "com.utmstack.crowdstrike"
	processName = "plugin_" + pluginName
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
)

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
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		_ = catcher.Error("ENCRYPTION_KEY is not set; the module configuration cannot be read", nil, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	err := modulecfg.Watch(context.Background(), os.Getenv("NATS_URL"), "crowdstrike", crypto.NewCipher(encKey),
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

const pluginKey = "crowdstrike"

var sensitiveKeys = map[string]bool{
	"crowdstrike_client_secret": true,
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
				Id:        id,
				GroupName: g.Name,
				TenantId:  tc.ID,
			}
			for k, v := range g.Config {
				conf := &Configuration{ConfKey: k, ConfValue: v}
				if encKey != "" && sensitiveKeys[conf.ConfKey] {
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
	sec.ModuleActive = len(sec.ModuleGroups) > 0
	return sec
}
