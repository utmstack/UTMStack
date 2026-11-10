package docker

import (
	"fmt"
	"os"
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
	Valkey        ValkeyConfig  `yaml:"valkey,omitempty"`
}

type ValkeyConfig struct {
	Addr     string `yaml:"addr,omitempty"`
	Password string `yaml:"password,omitempty"`
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

// feedsFile is the minimal shape for system_plugins_feeds.yaml. The existing
// PluginConfig carries many omitempty fields that would not marshal to the
// minimal shape, so a dedicated struct keeps the seed exact.
type feedsFile struct {
	Plugins map[string]feedsCfg `yaml:"plugins"`
}

type feedsCfg struct {
	Enabled bool `yaml:"enabled"`
}

func SetPluginsConfigs(conf *config.Config, stack *StackConfig) error {
	analysisPipeline := PluginsConfig{}
	analysisPipeline.Plugins = make(map[string]PluginConfig)
	analysisPipeline.Plugins["analysis"] = PluginConfig{
		Order: []string{"com.utmstack.events", "cel", "feeds", "com.utmstack.ad-audit", "com.utmstack.user-behavioral"},
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
		Valkey: ValkeyConfig{
			Addr:     "valkey:6379",
			Password: conf.Password,
		},
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

	feedsPath := filepath.Join(pipelineDir, "system_plugins_feeds.yaml")
	if err := seedFeedsConfig(feedsPath); err != nil {
		return fmt.Errorf("error seeding feeds plugin config: %w", err)
	}

	behavioralPath := filepath.Join(pipelineDir, "system_plugins_user_behavioral.yaml")
	if err := seedBehavioralConfig(behavioralPath); err != nil {
		return fmt.Errorf("error seeding user-behavioral plugin config: %w", err)
	}

	return nil
}

type behavioralFile struct {
	Plugins map[string]behavioralCfg `yaml:"plugins"`
}

type behavioralCfg struct {
	ZScoreLimit            float64 `yaml:"z_score_limit,omitempty"`
	ImpactMedium           int     `yaml:"impact_medium,omitempty"`
	ImpactHigh             int     `yaml:"impact_high,omitempty"`
	MaxConcurrentJudgments int     `yaml:"max_concurrent_judgments,omitempty"`
	CountCacheTTL          string  `yaml:"count_cache_ttl,omitempty"`
}

func seedBehavioralConfig(path string) error {
	var bf behavioralFile
	if _, err := os.Stat(path); err == nil {
		if err := utils.ReadYAML(path, &bf); err != nil {
			return fmt.Errorf("error reading %s: %w", path, err)
		}
	}
	if bf.Plugins == nil {
		bf.Plugins = map[string]behavioralCfg{}
	}

	changed := false
	if _, ok := bf.Plugins["user_behavioral"]; !ok {
		bf.Plugins["user_behavioral"] = behavioralCfg{
			ZScoreLimit:            3.0,
			ImpactMedium:           2,
			ImpactHigh:             3,
			MaxConcurrentJudgments: 32,
			CountCacheTTL:          "1m",
		}
		changed = true
	}

	if !changed {
		return nil
	}
	return utils.WriteYAML(path, bf)
}

func seedFeedsConfig(path string) error {
	var pf feedsFile
	if _, err := os.Stat(path); err == nil {
		if err := utils.ReadYAML(path, &pf); err != nil {
			return fmt.Errorf("error reading %s: %w", path, err)
		}
	}
	if pf.Plugins == nil {
		pf.Plugins = map[string]feedsCfg{}
	}

	changed := false
	if _, ok := pf.Plugins["feeds"]; !ok {
		pf.Plugins["feeds"] = feedsCfg{Enabled: true}
		changed = true
	}
	if _, ok := pf.Plugins["feeds_contribution"]; !ok {
		pf.Plugins["feeds_contribution"] = feedsCfg{Enabled: true}
		changed = true
	}

	if !changed {
		return nil
	}
	return utils.WriteYAML(path, pf)
}
