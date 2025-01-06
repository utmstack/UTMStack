package main

import (
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"time"

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
		utmConfig := plugins.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			_ = catcher.Error("internalKey or backendUrl is empty", nil, map[string]any{})
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		time.Sleep(delayCheckConfig)

		tempModuleConfig, err := client.GetUTMConfig(enum.GCP)
		if err != nil {
			if (err.Error() != "") && (err.Error() != " ") {
				_ = catcher.Error("cannot get GCP configuration", err, map[string]any{})
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
						_ = plugins.EnqueueNotification(plugins.TopicIntegrationFailure, plugins.Message{
							Id: uuid.NewString(),
							Message: catcher.Error("invalid credentials ", nil, map[string]any{
								"group": newConf.GroupName,
							}).Error(),
						})
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
