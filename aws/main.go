package main

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/aws/configuration"
	"github.com/utmstack/UTMStack/aws/processor"
	"github.com/utmstack/UTMStack/aws/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

func main() {
	catcher.Info("Starting aws module...", nil)
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
		if err := utils.ConnectionChecker(configuration.URL_CHECK_CONNECTION); err != nil {
			catcher.Error("Failed to establish connection", err, nil)
		}

		endTime := time.Now().UTC()

		catcher.Info("Syncing logs", map[string]any{"start": startTime, "end": endTime})

		moduleConfig, err := client.GetUTMConfig(enum.AWS_IAM_USER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				catcher.Error("error getting configuration of the AWS module: backend is not available", err, nil)
			}
			if strings.TrimSpace(err.Error()) != "" {
				catcher.Error("error getting configuration of the AWS module", err, nil)
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
							catcher.Error("program not configured yet for group", nil, map[string]any{"group": group.GroupName})
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

		catcher.Info("sync completed, waiting 5 minutes", map[string]any{"start": startTime, "end": endTime})
		startTime = endTime.Add(time.Nanosecond)
	}
}
