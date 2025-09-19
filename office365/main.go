package main

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/office365/configuration"
	"github.com/utmstack/UTMStack/office365/processor"
	"github.com/utmstack/UTMStack/office365/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

func main() {
	catcher.Info("Starting O365 module", map[string]any{})
	intKey := configuration.GetInternalKey()
	panelServ := configuration.GetPanelServiceName()
	if intKey == "" || panelServ == "" {
		catcher.Error("Internal key or panel service name is not set. Exiting...", nil, map[string]any{})
		os.Exit(1)
	}
	client := utmconf.NewUTMClient(intKey, "http://"+panelServ)

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	startTime := time.Now().UTC().Add(-delay)

	for range ticker.C {
		if err := utils.ConnectionChecker(configuration.LoginUrl); err != nil {
			catcher.Error("External connection failure detected", err, map[string]any{})
		}

		endTime := time.Now().UTC()

		catcher.Info("Syncing logs", map[string]any{"start_time": startTime, "end_time": endTime})

		moduleConfig, err := client.GetUTMConfig(enum.O365)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				catcher.Error("error getting configuration of the O365 module: backend is not available", nil, map[string]any{})
			}
			if strings.TrimSpace(err.Error()) != "" {
				catcher.Error("error getting configuration of the O365 module", err, map[string]any{})
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
							catcher.Error("program not configured yet for group", nil, map[string]any{"group_name": group.GroupName})
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

		catcher.Info("sync completed, waiting 5 minutes", map[string]any{"start_time": startTime, "end_time": endTime})
		startTime = endTime.Add(time.Nanosecond)
	}
}
