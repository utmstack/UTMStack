package main

import (
	"log"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/plugins/aws/config"
	"github.com/utmstack/UTMStack/plugins/aws/processor"
	"github.com/utmstack/UTMStack/plugins/aws/schema"
	"github.com/utmstack/UTMStack/plugins/aws/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
)

var (
	configs     = map[string]schema.AWSConfig{}
	newConfChan = make(chan struct{})
	logsChan    = make(chan *plugins.Log)
)

const delayCheckConfig = 30 * time.Second

func main() {
	pCfg, e := helpers.PluginCfg[schema.PluginConfig]("com.utmstack")
	if e != nil {
		log.Fatalf("failed to load plugin config: %v", e)
	}
	utils.InitLogger(pCfg.LogLevel)
	client := utmconf.NewUTMClient(pCfg.InternalKey, pCfg.Backend)

	go processor.ProcessLogs(logsChan)

	for {
		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.AWS_IAM_USER)
		if err != nil {
			utils.Logger.ErrorF("error getting configuration of the AWS module: %v", err)
			continue
		}

		if config.CompareConfigs(configs, tempModuleConfig.ConfigurationGroups) {
			utils.Logger.LogF(100, "configuration has been changed")
			close(newConfChan)
			newConfChan = make(chan struct{})
			configs = map[string]schema.AWSConfig{}

			for _, newConf := range tempModuleConfig.ConfigurationGroups {
				newConfiguration := schema.AWSConfig{
					AccessKeyID:     newConf.Configurations[0].ConfValue,
					SecretAccessKey: newConf.Configurations[1].ConfValue,
					Region:          newConf.Configurations[2].ConfValue,
					LogGroupName:    newConf.Configurations[3].ConfValue,
					LogStreamName:   newConf.Configurations[4].ConfValue,
				}
				configs[newConf.GroupName] = newConfiguration

				go processor.PullLogs(logsChan, newConfChan, newConf.GroupName, newConfiguration)
			}
		}
	}
}
