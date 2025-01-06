package main

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"os"
	"strings"
	"sync"
	"time"

	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	pCfg := plugins.PluginCfg("com.utmstack", false)
	internalKey := pCfg.Get("internalKey").String()
	backend := pCfg.Get("backend").String()

	client := utmconf.NewUTMClient(internalKey, backend)

	go plugins.SendLogsFromChannel()

	for {
		moduleConfig, err := client.GetUTMConfig(enum.SOPHOS)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				continue
			}
			_ = catcher.Error("cannot obtain module configuration", err, nil)
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
						skip = true
						break
					}
				}

				if !skip {
					pullLogs(group)
				}
				wg.Done()
			}(group)
		}

		wg.Wait()
		time.Sleep(time.Second * delayCheck)
	}
}
