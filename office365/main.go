package main

import (
	"strings"
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
	utils.Logger.Info("Starting O365 module")
	intKey := configuration.GetInternalKey()
	panelServ := configuration.GetPanelServiceName()
	client := utmconf.NewUTMClient(intKey, "http://"+panelServ)

	ticker := time.NewTicker(time.Second * delayCheck)
	defer ticker.Stop()

	startTime := time.Now().Add(-600 * time.Second).UTC()

	for range ticker.C {
		if err := utils.ConnectionChecker(configuration.LoginUrl); err != nil {
			utils.Logger.ErrorF("External connection failure detected: %v", err)
		}

		endTime := startTime.Add(299 * time.Second).UTC()
		startTimeStr := startTime.Format("2006-01-02T15:04:05")
		endTimeStr := endTime.Format("2006-01-02T15:04:05")

		utils.Logger.Info("syncing logs from %s to %s", startTime, endTime)

		moduleConfig, err := client.GetUTMConfig(enum.O365)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				startTime = endTime.Add(time.Second)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				utils.Logger.ErrorF("error getting configuration of the O365 module: %v", err)
			} else {
				utils.Logger.Info("program not configured yet")
			}

			startTime = endTime.Add(time.Second)

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
					processor.PullLogs(startTimeStr, endTimeStr, group)
				}

				wg.Done()
			}(group)
		}

		wg.Wait()

		utils.Logger.Info("sync complete waiting %d seconds", delayCheck)

		startTime = endTime.Add(time.Second)
	}
}
