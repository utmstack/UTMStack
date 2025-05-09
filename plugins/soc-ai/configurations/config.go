package configurations

import (
	"os"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
)

var (
	pluginConfig     PluginConfig
	pluginConfigOnce sync.Once
)

type PluginConfig struct {
	Backend                   string
	InternalKey               string
	APIKey                    string
	ChangeAlertStatus         bool
	AutomaticIncidentCreation bool
	Model                     string
	ModuleActive              bool
}

func GetPluginConfig() *PluginConfig {
	pluginConfigOnce.Do(func() {
		pluginConfig = PluginConfig{}
	})

	return &pluginConfig
}

func (c *PluginConfig) UpdateGPTConfigurations() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	for {
		if err := utils.ConnectionChecker(GPT_API_ENDPOINT); err != nil {
			_ = catcher.Error("Failed to establish internet connection: %v", err, nil)
		}

		utmConfig := plugins.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		moduleConfig, err := client.GetUTMConfig(enum.SOCAI)
		if err != nil && err.Error() != "" && err.Error() != " " {
			_ = catcher.Error("Error while getting GPT configuration: %v", err, nil)
			time.Sleep(TIME_FOR_GET_CONFIG * time.Second)
			continue
		}
		if moduleConfig == nil {
			time.Sleep(TIME_FOR_GET_CONFIG * time.Second)
			continue
		}

		c.Backend = backendUrl
		c.InternalKey = internalKey
		c.ModuleActive = moduleConfig.ModuleActive

		if c.ModuleActive && len(moduleConfig.ConfigurationGroups) > 0 {
			for _, config := range moduleConfig.ConfigurationGroups[0].Configurations {
				switch config.ConfKey {
				case "utmstack.socai.key":
					if config.ConfValue != "" && config.ConfValue != " " {
						c.APIKey = config.ConfValue
					}
				case "utmstack.socai.incidentCreation":
					if config.ConfValue != "" && config.ConfValue != " " {
						c.AutomaticIncidentCreation = config.ConfValue == "true"
					}
				case "utmstack.socai.changeAlertStatus":
					if config.ConfValue != "" && config.ConfValue != " " {
						c.ChangeAlertStatus = config.ConfValue == "true"
					}
				case "utmstack.socai.model":
					if config.ConfValue != "" && config.ConfValue != " " {
						c.Model = config.ConfValue
					}
				}
			}
		}

		time.Sleep(TIME_FOR_GET_CONFIG * time.Second)
	}
}
