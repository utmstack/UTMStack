package main

import (
	"log"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/utmstack/UTMStack/plugins/gcp/config"
	"github.com/utmstack/UTMStack/plugins/gcp/processor"
	"github.com/utmstack/UTMStack/plugins/gcp/schema"
	"github.com/utmstack/UTMStack/plugins/gcp/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
)

var (
	configs     = map[string]schema.ModuleConfig{}
	newConfChan = make(chan struct{})
	logsChan    = make(chan schema.LogEntry)
)

const delayCheckConfig = 30 * time.Second

func main() {
	pCfg, e := helpers.PluginCfg[schema.PluginConfig]("com.utmstack")
	if e != nil {
		log.Fatalf("Failed to load plugin config: %v", e)
	}
	utils.InitLogger(pCfg.LogLevel)
	client := utmconf.NewUTMClient(pCfg.InternalKey, pCfg.Backend)

	go processor.ProcessLogs(logsChan, *pCfg)

	for {
		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.GCP)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				utils.Logger.ErrorF("error getting configuration of the GCP module: %v", err)
			}
			continue
		}

		if config.CompareConfigs(configs, tempModuleConfig.ConfigurationGroups) {
			utils.Logger.Info("Configuration has been changed")
			close(newConfChan)
			newConfChan = make(chan struct{})
			configs = map[string]schema.ModuleConfig{}

			for _, newConf := range tempModuleConfig.ConfigurationGroups {
				newConfiguration := schema.ModuleConfig{
					ProjectID:      newConf.Configurations[1].ConfValue,
					SubscriptionID: newConf.Configurations[2].ConfValue,
					Topic:          newConf.Configurations[3].ConfValue,
					JsonKey:        newConf.Configurations[0].ConfValue,
				}
				configs[newConf.GroupName] = newConfiguration

				go processor.PullLogs(logsChan, newConfChan, newConf.GroupName, newConfiguration)
			}
		}
	}
}
