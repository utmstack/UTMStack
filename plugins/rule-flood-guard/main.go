package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/protobuf/types/known/emptypb"
)

const pluginID = "com.utmstack.rule-flood-guard"

func main() {
	initialCfg, configPath := loadConfig()

	osCfg := plugins.PluginCfg("org.opensearch").Get("opensearch")
	host := osCfg.Get("host").String()
	port := osCfg.Get("port").String()
	user := osCfg.Get("user").String()
	password := osCfg.Get("password").String()
	osURL := "https://" + host + ":" + port

	if err := sdkos.Connect([]string{osURL}, user, password); err != nil {
		_ = catcher.Error("rule-flood-guard: failed connecting to OpenSearch", err, map[string]any{"process": "plugin_" + pluginID})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

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
