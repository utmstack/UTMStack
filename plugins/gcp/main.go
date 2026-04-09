package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/gcp/config"
	"google.golang.org/api/option"
)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

type GroupModule struct {
	GroupName      string
	JsonKey        string
	ProjectID      string
	SubscriptionID string
	CTX            context.Context
	Cancel         context.CancelFunc
}

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.gcp").Env.Mode
	if mode != "worker" {
		return
	}

	go config.StartConfigurationSystem()

	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go plugins.SendLogsFromChannel("com.utmstack.gcp")
		go plugins.SendNotificationsFromChannel("com.utmstack.gcp")
	}

	startGroupModuleManager()

	// lock main until signal
	signs := make(chan os.Signal, 1)
	signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
	<-signs
}

func (g *GroupModule) PullLogs() {

	// Retry logic for creating client
	maxRetries := 3
	retryDelay := 2 * time.Second
	var client *pubsub.Client
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		client, err = pubsub.NewClient(g.CTX, g.ProjectID, option.WithCredentialsJSON([]byte(g.JsonKey)))
		if err == nil {
			break
		}

		_ = catcher.Error("failed to create client, retrying", err, map[string]any{
			"process":    "plugin_com.utmstack.gcp",
			"retry":      retry + 1,
			"maxRetries": maxRetries,
			"group":      g.GroupName,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	if err != nil {
		_ = catcher.Error("all retries failed when creating client", err, map[string]any{
			"process": "plugin_com.utmstack.gcp",
			"group":   g.GroupName,
		})
		return
	}

	defer func() { _ = client.Close() }()

	sub := client.Subscription(g.SubscriptionID)

	for {
		err := sub.Receive(g.CTX, func(ctx context.Context, msg *pubsub.Message) {
			plugins.EnqueueLog(&plugins.Log{
				Id:         uuid.NewString(),
				TenantId:   defaultTenant,
				DataType:   "google",
				DataSource: g.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        string(msg.Data),
			}, "com.utmstack.gcp")

			msg.Ack()
		})

		if err != nil {
			_ = catcher.Error("failed to receive message", err, map[string]any{"process": "plugin_com.utmstack.gcp"})
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

func getModuleConfig(newConf *config.ModuleGroup) GroupModule {
	gcpModule := GroupModule{}
	gcpModule.GroupName = newConf.GroupName
	gcpModule.CTX, gcpModule.Cancel = context.WithCancel(context.Background())
	for _, cnf := range newConf.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "jsonKey":
			gcpModule.JsonKey = cnf.ConfValue
		case "projectId":
			gcpModule.ProjectID = cnf.ConfValue
		case "subscription":
			gcpModule.SubscriptionID = cnf.ConfValue
		}
	}
	return gcpModule
}

type GroupModuleManager struct {
	Groups map[int32]GroupModule
}

func startGroupModuleManager() {
	manager := &GroupModuleManager{
		Groups: make(map[int32]GroupModule),
	}
	go manager.SyncConfigs()
}

func (m *GroupModuleManager) SyncConfigs() {
	time.Sleep(3 * time.Second)

	m.handleConfigUpdate(config.GetConfig())

	for newConfig := range config.GetConfigUpdateChannel() {
		catcher.Info("Received config update", map[string]any{
			"moduleActive": newConfig != nil && newConfig.ModuleActive,
			"process":      "plugin_com.utmstack.gcp",
		})
		m.handleConfigUpdate(newConfig)
	}
}

func (m *GroupModuleManager) handleConfigUpdate(moduleConfig *config.ConfigurationSection) {
	if err := ConnectionChecker(CHECKCON); err != nil {
		_ = catcher.Error("External connection failure detected", err, map[string]any{"process": "plugin_com.utmstack.gcp"})
	}

	if moduleConfig == nil || !moduleConfig.ModuleActive {
		for groupID, group := range m.Groups {
			catcher.Info("Cancelling group", map[string]any{
				"process": "plugin_com.utmstack.gcp",
			})
			group.Cancel()
			delete(m.Groups, groupID)
		}
		return
	}

	currentGroupIDs := make(map[int32]bool)
	for _, conf := range moduleConfig.ModuleGroups {
		currentGroupIDs[conf.Id] = true

		if existing, ok := m.Groups[conf.Id]; ok {
			newModule := getModuleConfig(conf)
			if configChanged(existing, newModule) {
				catcher.Info("Configuration changed for group, restarting", map[string]any{
					"process": "plugin_com.utmstack.gcp",
				})
				existing.Cancel()
				delete(m.Groups, conf.Id)
				m.Groups[conf.Id] = newModule
				go newModule.PullLogs()
			}
		} else {
			catcher.Info("Starting new group", map[string]any{
				"process": "plugin_com.utmstack.gcp",
			})
			m.Groups[conf.Id] = getModuleConfig(conf)
			group := m.Groups[conf.Id]
			go group.PullLogs()
		}
	}

	for groupID, group := range m.Groups {
		if !currentGroupIDs[groupID] {
			catcher.Info("Group removed, stopping", map[string]any{
				"process": "plugin_com.utmstack.gcp",
			})
			group.Cancel()
			delete(m.Groups, groupID)
		}
	}
}

func configChanged(old, new GroupModule) bool {
	return old.JsonKey != new.JsonKey ||
		old.ProjectID != new.ProjectID ||
		old.SubscriptionID != new.SubscriptionID
}
