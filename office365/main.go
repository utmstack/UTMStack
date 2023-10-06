package main

import (
	"sync"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/office365/configuration"
	"github.com/utmstack/UTMStack/office365/processor"
	UTMStackConfigurationClient "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

var h = holmes.New("debug", "O365_Integration")

const delayCheck = 300

func main() {
	st := time.Now().Add(-600 * time.Second)

	for {
		et := st.Add(299 * time.Second)
		startTime := st.UTC().Format("2006-01-02T15:04:05")
		endTime := et.UTC().Format("2006-01-02T15:04:05")

		intKey := configuration.GetInternalKey()
		panelServ := configuration.GetPanelServiceName()

		client := UTMStackConfigurationClient.NewUTMClient(intKey, "http://"+panelServ)
		moduleConfig, err := client.GetUTMConfig(enum.O365)
		if err != nil {
			if (err.Error() != "") && (err.Error() != " ") {
				h.Error("error getting configuration of the O365 module: %v", err)
			}
			h.Info("Sync complete waiting %v seconds", delayCheck)
			time.Sleep(time.Second * delayCheck)
			st = et.Add(1)
			continue
		}

		// If there is a configured tenant
		h.Info("Getting logs for groups")
		var wg sync.WaitGroup
		wg.Add(len(moduleConfig.ConfigurationGroups))
		for _, group := range moduleConfig.ConfigurationGroups {
			go func(group types.ModuleGroup) {
				for _, cnf := range group.Configurations {
					if cnf.ConfValue == "" || cnf.ConfValue == " " {
						h.Info("program not configured yet for group: %s", group.GroupName)
						continue
					}
				}
				processor.PullLogs(startTime, endTime, group, h)
				defer wg.Done()
			}(group)
		}

		wg.Wait()
		h.Info("Sync complete waiting %v seconds", delayCheck)
		time.Sleep(time.Second * delayCheck)
		st = et.Add(1)
	}
}
