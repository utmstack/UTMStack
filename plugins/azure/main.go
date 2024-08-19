package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

type PluginConfig struct {
	InternalKey string `yaml:"internal_key"`
	Backend     string `yaml:"backend"`
	LogLevel    int    `yaml:"log_level"`
}

const delayCheck = 300

func main() {
	pCfg, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		log.Fatalf("failed to load plugin config: %v", e)
	}
	initLogger(pCfg.LogLevel)
	client := utmconf.NewUTMClient(pCfg.InternalKey, pCfg.Backend)

	go processLogs()

	for {
		moduleConfig, err := client.GetUTMConfig(enum.AZURE)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				Logger.ErrorF("error getting configuration of the AZURE module: %v", err)
			}

			Logger.Info("sync complete waiting %v seconds", delayCheck)
			time.Sleep(time.Second * delayCheck)
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, group := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					var skip bool

					for _, cnf := range group.Configurations {
						if cnf.ConfValue == "" || cnf.ConfValue == " " {
							Logger.Info("program not configured yet for group: %s", group.GroupName)
							skip = true
							break
						}
					}

					if !skip {
						azureProcessor := getAzureProcessor(group)
						azureProcessor.pullLogs()
					}

					wg.Done()
				}(group)
			}

			wg.Wait()
			Logger.Info("sync complete waiting %d seconds", delayCheck)
		}

		time.Sleep(time.Second * delayCheck)
	}
}
