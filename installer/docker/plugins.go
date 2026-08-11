package docker

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

type PluginsConfig struct {
	Plugins map[string]PluginConfig `yaml:"plugins"`
}

type ClickHousePluginsConfig struct {
	Plugins map[string]ClickHouseConfig `yaml:"plugins"`
}

type PluginConfig struct {
	Order         []string      `yaml:"order,omitempty"`
	Port          int           `yaml:"port,omitempty"`
	PostgreSQL    PostgreConfig `yaml:"postgresql,omitempty"`
	InternalKey   string        `yaml:"internalKey,omitempty"`
	EncryptionKey string        `yaml:"encryptionKey,omitempty"`
	Env           string        `yaml:"env,omitempty"`
	AgentManager  string        `yaml:"agentManager,omitempty"`
	Backend       string        `yaml:"backend,omitempty"`
	CertsFolder   string        `yaml:"certsFolder,omitempty"`
}

type ClickHouseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`

	LogsTable       string `yaml:"logsTable"`
	AlertsTable     string `yaml:"alertsTable"`
	StatisticsTable string `yaml:"statisticsTable"`
}

type PostgreConfig struct {
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func SetPluginsConfigs(conf *config.Config, stack *StackConfig) error {
	analysisPipeline := PluginsConfig{}
	analysisPipeline.Plugins = make(map[string]PluginConfig)
	analysisPipeline.Plugins["analysis"] = PluginConfig{
		Order: []string{"com.utmstack.events", "cel", "feeds", "com.utmstack.ad-audit"},
	}

	correlationPipeline := PluginsConfig{}
	correlationPipeline.Plugins = make(map[string]PluginConfig)
	correlationPipeline.Plugins["correlation"] = PluginConfig{
		Order: []string{"com.utmstack.alerts", "com.utmstack.soc-ai", "com.utmstack.soar"},
	}

	notificationPipeline := PluginsConfig{}
	notificationPipeline.Plugins = make(map[string]PluginConfig)
	notificationPipeline.Plugins["notification"] = PluginConfig{
		Order: []string{"com.utmstack.stats"},
	}

	utmstackPipeline := PluginsConfig{}
	utmstackPipeline.Plugins = make(map[string]PluginConfig)
	utmstackPipeline.Plugins["com.utmstack"] = PluginConfig{
		PostgreSQL: PostgreConfig{
			Server:   "postgres",
			Port:     "5432",
			User:     "postgres",
			Password: conf.Password,
			Database: "utmstack",
		},
		InternalKey:   conf.InternalKey,
		EncryptionKey: conf.InternalKey,
		Env:           conf.Branch,
		AgentManager:  "agentmanager:9000",
		Backend:       "http://backend:8080",
		CertsFolder:   "/cert",
	}

	clickHousePipeline := ClickHousePluginsConfig{}
	clickHousePipeline.Plugins = make(map[string]ClickHouseConfig)
	clickHousePipeline.Plugins["clickhouse"] = ClickHouseConfig{
		Host:     "clickhouse",
		Port:     "9000",
		Database: "utmstack",
		User:     "default",
		Password: conf.Password,

		LogsTable:       "logs",
		AlertsTable:     "alerts",
		StatisticsTable: "statistics",
	}

	pipelineDir := filepath.Join(stack.EventsEngineWorkdir, "pipeline")
	utils.CreatePathIfNotExist(pipelineDir)

	err := utils.WriteYAML(filepath.Join(pipelineDir, "system_plugins_analysis.yaml"), analysisPipeline)
	if err != nil {
		return fmt.Errorf("error writing analysis pipeline config: %w", err)
	}

	err = utils.WriteYAML(filepath.Join(pipelineDir, "system_plugins_correlation.yaml"), correlationPipeline)
	if err != nil {
		return fmt.Errorf("error writing correlation pipeline config: %w", err)
	}

	err = utils.WriteYAML(filepath.Join(pipelineDir, "system_plugins_notification.yaml"), notificationPipeline)
	if err != nil {
		return fmt.Errorf("error writing notification pipeline config: %w", err)
	}

	err = utils.WriteYAML(filepath.Join(pipelineDir, "utmstack_plugins.yaml"), utmstackPipeline)
	if err != nil {
		return fmt.Errorf("error writing UTMStack pipeline config: %w", err)
	}

	err = utils.WriteYAML(filepath.Join(pipelineDir, "clickhouse_plugins.yaml"), clickHousePipeline)
	if err != nil {
		return fmt.Errorf("error writing ClickHouse pipeline config: %w", err)
	}

	return nil
}
