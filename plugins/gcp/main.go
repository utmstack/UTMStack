package main

import (
	"os"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/utmstack/UTMStack/plugins/gcp/config"
	"github.com/utmstack/UTMStack/plugins/gcp/processor"
	"github.com/utmstack/UTMStack/plugins/gcp/schema"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
)

var (
	configs     = map[string]schema.ModuleConfig{}
	newConfChan = make(chan struct{})
)

const delayCheckConfig = 30 * time.Second

func main() {
	pCfg, e := helpers.PluginCfg[schema.PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	client := utmconf.NewUTMClient(pCfg.InternalKey, pCfg.Backend)

	go processor.ProcessLogs()

	for {
		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.GCP)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				helpers.Logger().ErrorF("error getting configuration of the GCP module: %v", err)
			}
			continue
		}
		if tempModuleConfig.ModuleActive {
			if config.CompareConfigs(configs, tempModuleConfig.ConfigurationGroups) {
				helpers.Logger().LogF(100, "Configuration has been changed")
				close(newConfChan)
				newConfChan = make(chan struct{})
				configs = map[string]schema.ModuleConfig{}

				for _, newConf := range tempModuleConfig.ConfigurationGroups {
					newConfiguration := schema.ModuleConfig{
						JsonKey:        newConf.Configurations[0].ConfValue,
						ProjectID:      newConf.Configurations[1].ConfValue,
						SubscriptionID: newConf.Configurations[2].ConfValue,
						Topic:          newConf.Configurations[3].ConfValue,
					}
					configs[newConf.GroupName] = newConfiguration

					go processor.PullLogs(newConfChan, newConf.GroupName, newConfiguration)
				}
			}
		}
	}
}
