package main

import (
	"os"
	"strings"
	"sync"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/plugins/sophos/processor"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

type PluginConfig struct {
	InternalKey string `yaml:"internalKey"`
	Backend     string `yaml:"backend"`
}

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

	go processor.ProcessLogs()

	for {
		moduleConfig, err := client.GetUTMConfig(enum.SOPHOS)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				go_sdk.Logger().LogF(100, "error getting configuration of the SOPHOS module: invalid character '<'")
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				go_sdk.Logger().ErrorF("error getting configuration of the SOPHOS module: %v", err)
			}

			go_sdk.Logger().Info("sync complete waiting %v seconds", delayCheck)
			time.Sleep(delayCheck * time.Second)
			continue
		}

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
					processor.PullLogs(group)
				}
				wg.Done()
			}(group)
		}

		wg.Wait()
		go_sdk.Logger().Info("sync complete waiting %d seconds", delayCheck)
		time.Sleep(time.Second * delayCheck)
	}
}
