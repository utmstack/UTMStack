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

func main() {
	utils.Logger.Info("Starting aws module...")
	intKey := configuration.GetInternalKey()
	panelServ := configuration.GetPanelServiceName()
	if intKey == "" || panelServ == "" {
		utils.Logger.Fatal("Internal key or panel service name is not set. Exiting...")
	}
	client := utmconf.NewUTMClient(intKey, "http://"+panelServ)

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	startTime := time.Now().UTC().Add(-delay)

	for range ticker.C {
		if err := utils.ConnectionChecker(configuration.URL_CHECK_CONNECTION); err != nil {
			utils.Logger.ErrorF("Failed to establish connection: %v", err)
		}

		endTime := time.Now().UTC()

		utils.Logger.Info("Syncing logs from %s to %s", startTime, endTime)

		moduleConfig, err := client.GetUTMConfig(enum.AWS_IAM_USER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				utils.Logger.LogF(100, "error getting configuration of the AWS module: backend is not available")
			}
			if strings.TrimSpace(err.Error()) != "" {
				utils.Logger.ErrorF("error getting configuration of the AWS module: %v", err)
			}
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, grp := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					defer wg.Done()
					var skip bool

					for _, cnf := range group.Configurations {
						if strings.TrimSpace(cnf.ConfValue) == "" {
							utils.Logger.LogF(100, "program not configured yet for group: %s", group.GroupName)
							skip = true
							break
						}
					}

					if !skip {
						processor.PullLogs(startTime, endTime, group)
					}
				}(grp)
			}
			wg.Wait()
		}

		utils.Logger.Info("sync completed from %v to %v, waiting 5 minutes", startTime, endTime)
		startTime = endTime.Add(time.Nanosecond)
	}
}
