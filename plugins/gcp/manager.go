package main

import (
	"strings"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
)

const delayCheckConfig = 30 * time.Second

type GroupModuleManager struct {
	Groups map[int]GroupModule
}

func StartGroupModuleManager() {
	manager := &GroupModuleManager{
		Groups: make(map[int]GroupModule),
	}
	go manager.SyncConfigs()
}

func (m *GroupModuleManager) SyncConfigs() {
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

		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.GCP)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				go_sdk.Logger().LogF(100, "error getting configuration of the GCP module: %v", err)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				go_sdk.Logger().ErrorF("error getting configuration of the GCP module: %v", err)
			}
			continue
		}
		if tempModuleConfig.ModuleActive {
			for _, newConf := range tempModuleConfig.ConfigurationGroups {
				cnfChangedOrNew := false
				if _, ok := m.Groups[newConf.ID]; !ok {
					cnfChangedOrNew = true
				} else {
					cnfChangedOrNew = compareConfigs(m.Groups[newConf.ID], newConf)
					if cnfChangedOrNew {
						m.Groups[newConf.ID].Cancel()
					}
				}
				if cnfChangedOrNew {
					go_sdk.Logger().LogF(100, "Checking GCP group configuration: %s", newConf.GroupName)
					m.Groups[newConf.ID] = GetModuleConfig(newConf)
					group := m.Groups[newConf.ID]
					isValid := group.VerifyCredentials()
					if !isValid {
						go_sdk.Logger().ErrorF("Invalid credentials for group %s", newConf.GroupName)
						notify("gpc_invalid_credentials", Message{Cause: "Invalid credentials for group " + newConf.GroupName, DataType: "gcp", DataSource: newConf.GroupName})
						continue
					}
					go group.PullLogs()
				}
			}
		} else {
			go_sdk.Logger().LogF(100, "GCP module is disabled")
			for _, cnf := range m.Groups {
				cnf.Cancel()
			}
		}
	}
}
