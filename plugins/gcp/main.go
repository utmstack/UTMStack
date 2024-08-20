package main

import (
	"os"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
)

var (
	configs     = map[string]ModuleConfig{}
	newConfChan = make(chan struct{})
)

const delayCheckConfig = 30 * time.Second

func main() {
	helpers.Logger().Info("Starting GCP plugin")
	pCfg, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	client := utmconf.NewUTMClient(pCfg.InternalKey, pCfg.Backend)

	go processLogs()

	for {
		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.GCP)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				helpers.Logger().LogF(100, "error getting configuration of the GCP module: %v", err)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				helpers.Logger().ErrorF("error getting configuration of the GCP module: %v", err)
			}
			continue
		}
		if tempModuleConfig.ModuleActive {
			if compareConfigs(configs, tempModuleConfig.ConfigurationGroups) {
				helpers.Logger().LogF(100, "Configuration has been changed")
				close(newConfChan)
				newConfChan = make(chan struct{})
				configs = map[string]ModuleConfig{}

				for _, newConf := range tempModuleConfig.ConfigurationGroups {
					newConfiguration := ModuleConfig{
						JsonKey:        newConf.Configurations[0].ConfValue,
						ProjectID:      newConf.Configurations[1].ConfValue,
						SubscriptionID: newConf.Configurations[2].ConfValue,
						Topic:          newConf.Configurations[3].ConfValue,
					}
					configs[newConf.GroupName] = newConfiguration

					go pullLogs(newConfChan, newConf.GroupName, newConfiguration)
				}
			}
		} else {
			helpers.Logger().LogF(100, "GCP module is disabled")
		}
	}
}
