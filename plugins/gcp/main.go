package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/api/option"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"
const delayCheckConfig = 30 * time.Second

type GroupModule struct {
	GroupName      string
	JsonKey        string
	ProjectID      string
	SubscriptionID string
	CTX            context.Context
	Cancel         context.CancelFunc
}

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "worker" {
		os.Exit(0)
	}

	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go plugins.SendLogsFromChannel()
		go plugins.SendNotificationsFromChannel()
	}

	startGroupModuleManager()

	// lock main until signal
	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	<-signs
}

func (g *GroupModule) PullLogs() {
	client, err := pubsub.NewClient(g.CTX, g.ProjectID, option.WithCredentialsJSON([]byte(g.JsonKey)))
	if err != nil {
		_ = catcher.Error("failed to create client", err, map[string]any{})
		return
	}

	defer func() { _ = client.Close() }()

	sub := client.Subscription(g.SubscriptionID)

	for {

		err = sub.Receive(g.CTX, func(ctx context.Context, msg *pubsub.Message) {
			plugins.EnqueueLog(&plugins.Log{
				Id:         uuid.NewString(),
				TenantId:   defaultTenant,
				DataType:   "google",
				DataSource: g.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        string(msg.Data),
			})

			msg.Ack()
		})

		if err != nil {
			_ = catcher.Error("failed to receive message", err, map[string]any{})
			continue
		}
	}
}

func (g *GroupModule) VerifyCredentials() bool {
	client, err := pubsub.NewClient(g.CTX, g.ProjectID, option.WithCredentialsJSON([]byte(g.JsonKey)))
	if err != nil {
		_ = catcher.Error("failed to create client", err, map[string]any{})
		return false
	}
	defer func() { _ = client.Close() }()

	sub := client.Subscription(g.SubscriptionID)
	exist, err := sub.Exists(g.CTX)
	if err != nil {
		_ = catcher.Error("failed to verify subscription", err, map[string]any{})
		return false
	}
	if !exist {
		_ = catcher.Error("subscription does not exist", nil, map[string]any{})
		return false
	}
	return true
}

func getModuleConfig(newConf types.ModuleGroup) GroupModule {
	ctx, cancel := context.WithCancel(context.Background())
	return GroupModule{
		GroupName:      newConf.GroupName,
		JsonKey:        newConf.Configurations[0].ConfValue,
		ProjectID:      newConf.Configurations[1].ConfValue,
		SubscriptionID: newConf.Configurations[2].ConfValue,
		CTX:            ctx,
		Cancel:         cancel,
	}
}

func compareConfigs(saveConf GroupModule, newConf types.ModuleGroup) (isNecessaryVerify bool) {
	if saveConf.JsonKey != newConf.Configurations[0].ConfValue ||
		saveConf.ProjectID != newConf.Configurations[1].ConfValue ||
		saveConf.SubscriptionID != newConf.Configurations[2].ConfValue {
		return true
	}
	return false
}

type GroupModuleManager struct {
	Groups map[int]GroupModule
}

func startGroupModuleManager() {
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
					m.Groups[newConf.ID] = getModuleConfig(newConf)
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
