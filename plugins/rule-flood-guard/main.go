package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store"
	ch "github.com/threatwinds/go-sdk/store/clickhouse"
	"google.golang.org/protobuf/types/known/emptypb"
)

const pluginID = "com.utmstack.rule-flood-guard"

const (
	datasetAlerts      = store.Dataset("alerts")
	defaultAlertsTable = "alerts"
)

var alertStore *ch.Driver

func main() {
	initialCfg, configPath := loadConfig()

	cfg := plugins.PluginCfg("clickhouse")
	table := cfg.Get("alertsTable").String()
	if table == "" {
		table = defaultAlertsTable
	}

	var err error
	alertStore, err = ch.New(ch.Config{
		Addr:     []string{cfg.Get("host").String() + ":" + cfg.Get("port").String()},
		Database: cfg.Get("database").String(),
		Username: cfg.Get("user").String(),
		Password: cfg.Get("password").String(),
		Tables: map[store.Dataset]string{
			datasetAlerts: table,
		},
		TenantColumn:   "tenantId",
		TimeColumn:     "@timestamp",
		DataTypeColumn: "dataType",
	})
	if err != nil {
		_ = catcher.Error("rule-flood-guard: cannot build the alert store", err, map[string]any{"process": "plugin_" + pluginID})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	defer alertStore.Close()

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
