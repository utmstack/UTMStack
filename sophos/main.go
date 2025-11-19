package main

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/sophos/configuration"
	"github.com/utmstack/UTMStack/sophos/processor"
	"github.com/utmstack/UTMStack/sophos/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

func main() {
	catcher.Info("Starting sophos central module...", nil)
	intKey := configuration.GetInternalKey()
	panelServ := configuration.GetPanelServiceName()
	if intKey == "" || panelServ == "" {
		catcher.Error("Internal key or panel service name is not set. Exiting...", nil, nil)
		os.Exit(1)
	}
	client := utmconf.NewUTMClient(intKey, "http://"+panelServ)

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	startTime := time.Now().UTC().Add(-delay)

	for range ticker.C {
		if err := utils.ConnectionChecker(configuration.CHECKCON); err != nil {
			catcher.Error("External connection failure detected", err, nil)
		}

		endTime := time.Now().UTC()
		//startTimeStr := startTime.Format(time.RFC3339)
		//endTimeStr := endTime.Format(time.RFC3339)

		catcher.Info("Syncing logs", map[string]any{"start_time": startTime, "end_time": endTime})

		moduleConfig, err := client.GetUTMConfig(enum.SOPHOS)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				catcher.Error("error getting configuration of the SOPHOS module: backend is not available", nil, nil)
			}
			if strings.TrimSpace(err.Error()) != "" {
				catcher.Error("error getting configuration of the SOPHOS module", err, nil)
			}
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, grp := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					var skip bool

					for _, cnf := range group.Configurations {
						if strings.TrimSpace(cnf.ConfValue) == "" {
							catcher.Error("program not configured yet for group", nil, map[string]any{"group_name": group.GroupName})
							skip = true
							break
						}
					}

					if !skip {
						processor.PullLogs(group, startTime)
					}

					wg.Done()
				}(grp)
			}
			wg.Wait()
		}

		catcher.Info("sync completed, waiting 5 minutes", map[string]any{"start_time": startTime, "end_time": endTime})
		startTime = endTime.Add(time.Nanosecond)
	}
}
