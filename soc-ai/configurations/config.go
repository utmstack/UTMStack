package configurations

import (
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	UTMStackConfigurationClient "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/soc-ai/utils"
)

var (
	gptConfig     GPTConfig
	gptConfigOnce sync.Once
)

type GPTConfig struct {
	APIKey                    string
	ChangeAlertStatus         bool
	AutomaticIncidentCreation bool
	Model                     string
	ModuleActive              bool
}

func GetGPTConfig() *GPTConfig {
	gptConfigOnce.Do(func() {
		gptConfig = GPTConfig{}
	})

	return &gptConfig
}

func (c *GPTConfig) UpdateGPTConfigurations() {
	intKey := GetInternalKey()
	panelServ := GetPanelServiceName()
	client := UTMStackConfigurationClient.NewUTMClient(intKey, panelServ)

	catcher.Info("Starting to update GPT configurations...", nil)

	for {
		if err := utils.ConnectionChecker(GPT_API_ENDPOINT); err != nil {
			catcher.Error("Failed to establish internet connection", err, nil)
		}

		tempModuleConfig, err := client.GetUTMConfig(enum.SOCAI)
		if err != nil && err.Error() != "" && err.Error() != " " {
			catcher.Error("Error while getting GPT configuration", err, nil)
			time.Sleep(TIME_FOR_GET_CONFIG * time.Second)
			continue
		}
		if tempModuleConfig == nil {
			catcher.Error("Got nil config from server", nil, nil)
			time.Sleep(TIME_FOR_GET_CONFIG * time.Second)
			continue
		}

		c.ModuleActive = tempModuleConfig.ModuleActive

		if c.ModuleActive && len(tempModuleConfig.ConfigurationGroups) > 0 {
			for _, config := range tempModuleConfig.ConfigurationGroups[0].Configurations {
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
