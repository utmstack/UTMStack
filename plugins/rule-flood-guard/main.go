package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store"
	"github.com/threatwinds/go-sdk/store/factory"
	"github.com/utmstack/UTMStack/plugins/rule-flood-guard/internal/storecfg"
	"google.golang.org/protobuf/types/known/emptypb"
)

const pluginID = "com.utmstack.rule-flood-guard"

const (
	datasetAlerts      = store.Dataset("alerts")
	defaultAlertsTable = "alerts"
)

var alertStore store.Store

func main() {
	initialCfg, configPath := loadConfig()

	cfg, err := storecfg.New(plugins.PluginCfg("store"), plugins.PluginCfg("clickhouse"), map[string]storecfg.Table{
		string(datasetAlerts): {ConfigKey: "alertsTable", Default: defaultAlertsTable},
	})
	if err == nil {
		cfg, err = storecfg.ClickHousePrimary(cfg)
	}
	if err != nil {
		_ = catcher.Error("rule-flood-guard: invalid alert store configuration", err, map[string]any{"process": "plugin_" + pluginID})
		os.Exit(1)
	}
	built, err := factory.New(cfg, factory.Options{})
	if err != nil {
		_ = catcher.Error("rule-flood-guard: cannot build the alert store", err, map[string]any{"process": "plugin_" + pluginID})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	defer built.Close()
	alertStore = built.Store

	client := newBackendClient(initialCfg.BackendURL, initialCfg.InternalKey)
	holder := newConfigHolder(initialCfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watchConfigFile(ctx, holder, configPath)
	go runLoop(ctx, searchRuleBuckets, client, holder.Get)

	if err := plugins.InitNotificationPlugin(pluginID, notify); err != nil {
		_ = catcher.Error("rule-flood-guard: failed to start notification plugin", err, map[string]any{"process": "plugin_" + pluginID})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func notify(_ context.Context, _ *plugins.Message) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
