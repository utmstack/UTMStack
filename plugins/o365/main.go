package main

import (
	"os"
	"strings"
	"sync"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/plugins/o365/processor"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

func main() {
	pCfg, e := go_sdk.PluginCfg[processor.PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	client := utmconf.NewUTMClient(pCfg.InternalKey, pCfg.Backend)

	go processor.ProcessLogs()

	st := time.Now().Add(-delayCheck * time.Second)

	for {
		et := st.Add(299 * time.Second)
		startTime := st.UTC().Format("2006-01-02T15:04:05")
		endTime := et.UTC().Format("2006-01-02T15:04:05")

		go_sdk.Logger().Info("syncing logs from %s to %s", startTime, endTime)

		moduleConfig, err := client.GetUTMConfig(enum.O365)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				go_sdk.Logger().LogF(100, "error getting configuration of the O365 module: %v", err)
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				go_sdk.Logger().ErrorF("error getting configuration of the O365 module: %v", err)
			} else {
				go_sdk.Logger().Info("program not configured yet")
			}

			time.Sleep(time.Second * delayCheck)
			st = time.Now().Add(-delayCheck * time.Second)
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, group := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					go_sdk.Logger().Info("getting logs for group: %s", group.GroupName)
					var skip bool

					for _, cnf := range group.Configurations {
						if cnf.ConfValue == "" || cnf.ConfValue == " " {
							go_sdk.Logger().Info("program not configured yet for group: %s", group.GroupName)
							skip = true
							break
						}
					}

					if !skip {
						processor.PullLogs(startTime, endTime, group)
					}

					wg.Done()
				}(group)
			}

			wg.Wait()
		} else {
			go_sdk.Logger().LogF(100, "module O365 is not active")
		}

		go_sdk.Logger().Info("sync complete waiting %d seconds", delayCheck)
		time.Sleep(time.Second * delayCheck)
		st = time.Now().Add(-delayCheck * time.Second)
	}
}
