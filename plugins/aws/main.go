package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/helpers"

	"github.com/utmstack/UTMStack/plugins/aws/processor"
	"github.com/utmstack/UTMStack/plugins/aws/schema"
	"github.com/utmstack/UTMStack/plugins/aws/utils"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

var (
	configs     = map[string]schema.AWSConfig{}
	newConfChan = make(chan struct{})
)

const delayCheck = 300

func main() {
	pCfg, e := helpers.PluginCfg[schema.PluginConfig]("com.utmstack")
	if e != nil {
		log.Fatalf("failed to load plugin config: %v", e)
	}
	utils.InitLogger(pCfg.LogLevel)
	client := utmconf.NewUTMClient(pCfg.InternalKey, pCfg.Backend)

	st := time.Now().Add(-600 * time.Second)
	go processor.ProcessLogs()

	for {
		et := st.Add(299 * time.Second)
		moduleConfig, err := client.GetUTMConfig(enum.AWS_IAM_USER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
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

		if moduleConfig.ModuleActive {
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
		}

		time.Sleep(time.Second * delayCheck)
		st = et.Add(1)
	}
}
