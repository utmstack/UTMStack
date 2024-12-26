package main

import (
	"time"

	gosdk "github.com/threatwinds/go-sdk"
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
		utmConfig := gosdk.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			_ = gosdk.Error("internalKey or backendUrl is empty", nil, map[string]any{})
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.GCP)
		if err != nil {
			if (err.Error() != "") && (err.Error() != " ") {
				_ = gosdk.Error("cannot get GCP configuration", err, map[string]any{})
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
					m.Groups[newConf.ID] = GetModuleConfig(newConf)
					group := m.Groups[newConf.ID]
					isValid := group.VerifyCredentials()
					if !isValid {
						_ = gosdk.Error("invalid credentials for group "+newConf.GroupName, nil, map[string]any{})
						notify("gpc_invalid_credentials", Message{Cause: "Invalid credentials for group " + newConf.GroupName, DataType: "gcp", DataSource: newConf.GroupName})
						continue
					}
					go group.PullLogs()
				}
			}
		} else {
			for _, cnf := range m.Groups {
				cnf.Cancel()
			}
		}
	}
}
