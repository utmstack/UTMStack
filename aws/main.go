package main

import (
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/aws/configuration"
	"github.com/utmstack/UTMStack/aws/processor"
	"github.com/utmstack/UTMStack/aws/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

func main() {
	utils.Logger.Info("Starting aws module...")
	intKey := configuration.GetInternalKey()
	panelServ := configuration.GetPanelServiceName()
	client := utmconf.NewUTMClient(intKey, "http://"+panelServ)

	st := time.Now().Add(-600 * time.Second)

	for {
		if err := utils.ConnectionChecker(configuration.URL_CHECK_CONNECTION); err != nil {
			utils.Logger.ErrorF("Failed to establish connection: %v", err)
		}

		et := st.Add(299 * time.Second)
		moduleConfig, err := client.GetUTMConfig(enum.AWS_IAM_USER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				utils.Logger.LogF(100, "error getting configuration of the AWS module: %v", err)
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				utils.Logger.ErrorF("error getting configuration of the AWS module: %v", err)
			}

			utils.Logger.Info("sync complete waiting %v seconds", delayCheck)
			time.Sleep(time.Second * delayCheck)
			st = et.Add(1)
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
					processor.PullLogs(st, et, group)
				}

				wg.Done()
			}(group)
		}

		wg.Wait()
		utils.Logger.Info("sync complete waiting %d seconds", delayCheck)
		time.Sleep(time.Second * delayCheck)
		st = et.Add(1)
	}
}
