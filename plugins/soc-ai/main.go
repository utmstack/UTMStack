package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/agent"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/api"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/queue"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	if plugins.GetCfg("plugin_com.utmstack.soc-ai").GetEnv().Mode == "playground" {
		return
	}

	go config.StartConfigurationSystem()
	go applyConfigUpdates()

	time.Sleep(config.CONFIG_STARTUP_DELAY * time.Second)
	queue.Initialize()

	// HTTP API: manual submission + live agent operations (SSE).
	go api.StartHTTPServer()

	err := plugins.InitCorrelationPlugin("com.utmstack.soc-ai", correlate)
	if err != nil {
		_ = catcher.Error("failed to start correlation plugin", err, map[string]any{
			"process": "plugin_com.utmstack.soc-ai",
		})
		time.Sleep(config.ERROR_EXIT_DELAY * time.Second)
		os.Exit(1)
	}
}

func applyConfigUpdates() {
	for set := range config.GetConfigUpdateChannel() {
		if set == nil || set.Default == nil {
			agent.SetBuilder(nil)
			continue
		}

		agent.SetBuilder(func(tenantID string) *agent.Agent {
			cfg := config.GetConfig(tenantID)
			if cfg == nil || !cfg.ModuleActive {
				return nil
			}

			llm := agent.NewLLMClient(agent.LLMConfig{
				Provider:       cfg.Provider,
				URL:            cfg.URL,
				Model:          cfg.Model,
				APIKey:         cfg.APIKey,
				AuthType:       cfg.AuthType,
				AuthHeaderName: cfg.AuthHeaderName,
				CustomHeaders:  cfg.CustomHeaders,
			})
			broker := agent.NewToolBroker(cfg.Backend, cfg.InternalKey, tenantID)

			catcher.Info("SOC-AI agent configured", map[string]any{
				"process":  "plugin_com.utmstack.soc-ai",
				"tenant":   tenantID,
				"provider": cfg.Provider,
				"model":    cfg.Model,
			})

			return agent.New(llm, broker, cfg.Model, cfg.MaxTokens, cfg.ContextWindow)
		})

		if !set.Default.ModuleActive && len(set.Tenants) == 0 {
			catcher.Info("SOC-AI inactive (incomplete configuration)", map[string]any{"process": "plugin_com.utmstack.soc-ai"})
		}
	}
}

func correlate(_ context.Context, alert *plugins.Alert) (*emptypb.Empty, error) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in Correlate method", nil, map[string]any{
				"process": "plugin_com.utmstack.soc-ai",
				"panic":   r,
				"alert":   alert.Name,
			})
		}
	}()

	cfg := config.GetConfig(alert.TenantId)
	if cfg == nil || !cfg.ModuleActive || !cfg.AutoAnalyze {
		return &emptypb.Empty{}, nil
	}
	if agent.For(alert.TenantId) == nil {
		return &emptypb.Empty{}, nil
	}

	queue.Enqueue(alert)
	return &emptypb.Empty{}, nil
}
