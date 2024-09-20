package main

import (
	"os"
	"strings"
	"sync"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

type PluginConfig struct {
	InternalKey string `yaml:"internalKey"`
	Backend     string `yaml:"backend"`
}

const delayCheck = 300

func main() {
	mode := os.Getenv("MODE")
	if mode != "manager" {
		os.Exit(0)
	}
	
	pCfg, e := go_sdk.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

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
				go_sdk.Logger().ErrorF("error getting configuration of the AZURE module: %v", err)
			}

			go_sdk.Logger().Info("sync complete waiting %v seconds", delayCheck)
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
							go_sdk.Logger().Info("program not configured yet for group: %s", group.GroupName)
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
			go_sdk.Logger().Info("sync complete waiting %d seconds", delayCheck)
		}

		time.Sleep(time.Second * delayCheck)
	}
}
