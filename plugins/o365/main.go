package main

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/plugins/o365/processor"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "worker" {
		os.Exit(0)
	}

	go processor.ProcessLogs()

	st := time.Now().Add(-delayCheck * time.Second)

	for {
		pConfig := plugins.PluginCfg("com.utmstack", false)
		backend := pConfig.Get("backend").String()
		internalKey := pConfig.Get("internalKey").String()

		client := utmconf.NewUTMClient(internalKey, backend)

		et := st.Add(299 * time.Second)
		startTime := st.UTC().Format("2006-01-02T15:04:05")
		endTime := et.UTC().Format("2006-01-02T15:04:05")

		moduleConfig, err := client.GetUTMConfig(enum.O365)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				_ = catcher.Error("cannot obtain module configuration", err, nil)
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
					var skip bool

					for _, cnf := range group.Configurations {
						if cnf.ConfValue == "" || cnf.ConfValue == " " {
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
		}

		time.Sleep(time.Second * delayCheck)
		st = time.Now().Add(-delayCheck * time.Second)
	}
}
