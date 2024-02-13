package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/office365/configuration"
	"github.com/utmstack/UTMStack/office365/processor"
	"github.com/utmstack/UTMStack/office365/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

func main() {
	intKey := configuration.GetInternalKey()
	panelServ := configuration.GetPanelServiceName()
	client := utmconf.NewUTMClient(intKey, "http://"+panelServ)

	st := time.Now().Add(-600 * time.Second)

	for {
		et := st.Add(299 * time.Second)
		startTime := st.UTC().Format("2006-01-02T15:04:05")
		endTime := et.UTC().Format("2006-01-02T15:04:05")

		moduleConfig, err := client.GetUTMConfig(enum.O365)
		if err != nil {
			if (err.Error() != "") && (err.Error() != " ") {
				utils.Logger.ErrorF(http.StatusInternalServerError, "error getting configuration of the O365 module: %v", err)
			}

			utils.Logger.Info("sync complete waiting %v seconds", delayCheck)

			time.Sleep(time.Second * delayCheck)

			st = et.Add(1)

			continue
		}

		utils.Logger.Info("getting logs for groups")

		var wg sync.WaitGroup
		
		wg.Add(len(moduleConfig.ConfigurationGroups))

		for _, group := range moduleConfig.ConfigurationGroups {
			go func(group types.ModuleGroup) {
				var skip bool
				
				for _, cnf := range group.Configurations {
					if cnf.ConfValue == "" || cnf.ConfValue == " " {
						utils.Logger.Info("program not configured yet for group: %s", group.GroupName)
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

		utils.Logger.Info("waiting %d seconds until sync completes", delayCheck)

		wg.Wait()

		utils.Logger.Info("sync complete waiting %d seconds", delayCheck)

		time.Sleep(time.Second * delayCheck)

		st = et.Add(1)
	}
}
