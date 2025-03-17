package main

import (
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/sophos/configuration"
	"github.com/utmstack/UTMStack/sophos/processor"
	"github.com/utmstack/UTMStack/sophos/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

func main() {
	intKey := configuration.GetInternalKey()
	panelServ := configuration.GetPanelServiceName()
	client := utmconf.NewUTMClient(intKey, "http://"+panelServ)

	for {
		if err := utils.ConnectionChecker(configuration.CHECKCON); err != nil {
			utils.Logger.ErrorF("External connection failure detected: %v", err)
		}

		moduleConfig, err := client.GetUTMConfig(enum.SOPHOS)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				utils.Logger.ErrorF("error getting configuration of the SOPHOS module: %v", err)
			}

			utils.Logger.Info("sync complete waiting %v seconds", delayCheck)
			time.Sleep(time.Second * delayCheck)
			continue
		}

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
					processor.PullLogs(group)
				}

				wg.Done()
			}(group)
		}

		wg.Wait()
		utils.Logger.Info("sync complete waiting %d seconds", delayCheck)
		time.Sleep(time.Second * delayCheck)
	}
}
