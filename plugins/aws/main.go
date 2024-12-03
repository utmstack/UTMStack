package main

import (
	"os"
	"strings"
	"sync"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/plugins/aws/processor"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

func main() {
	mode := go_sdk.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	st := time.Now().Add(-600 * time.Second)
	go processor.ProcessLogs()

	for {
		utmConfig := go_sdk.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			go_sdk.Logger().ErrorF("internalKey or backendUrl is empty")
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		et := st.Add(299 * time.Second)
		moduleConfig, err := client.GetUTMConfig(enum.AWS_IAM_USER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				go_sdk.Logger().ErrorF("error getting configuration of the AWS module: %v", err)
			}

			go_sdk.Logger().Info("sync complete waiting %v seconds", delayCheck)
			time.Sleep(time.Second * delayCheck)
			st = et.Add(1)
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
							go_sdk.Logger().Info("program not configured yet for group: %s", group.GroupName)
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
			go_sdk.Logger().Info("sync complete waiting %d seconds", delayCheck)
		}

		time.Sleep(time.Second * delayCheck)
		st = et.Add(1)
	}
}
