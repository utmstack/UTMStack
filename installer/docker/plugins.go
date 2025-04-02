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

type PluginConfig struct {
	Order         []string      `yaml:"order,omitempty"`
	Port          int           `yaml:"port,omitempty"`
	RulesFolder   string        `yaml:"rulesFolder,omitempty"`
	GeoIPFolder   string        `yaml:"geoipFolder,omitempty"`
	Elasticsearch string        `yaml:"elasticsearch,omitempty"`
	PostgreSQL    PostgreConfig `yaml:"postgresql,omitempty"`
	ServerName    string        `yaml:"serverName,omitempty"`
	InternalKey   string        `yaml:"internalKey,omitempty"`
	AgentManager  string        `yaml:"agentManager,omitempty"`
	Backend       string        `yaml:"backend,omitempty"`
	CertsFolder   string        `yaml:"certsFolder,omitempty"`
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
		Order: []string{"com.utmstack.events", "cel"},
	}

	correlationPipeline := PluginsConfig{}
	correlationPipeline.Plugins = make(map[string]PluginConfig)
	correlationPipeline.Plugins["correlation"] = PluginConfig{
		Order: []string{"com.utmstack.events"},
	}

	inputPipeline := PluginsConfig{}
	inputPipeline.Plugins = make(map[string]PluginConfig)
	inputPipeline.Plugins["http-input"] = PluginConfig{
		Port: 8082,
	}
	inputPipeline.Plugins["grpc-input"] = PluginConfig{
		Port: 8083,
	}

	notificationPipeline := PluginsConfig{}
	notificationPipeline.Plugins = make(map[string]PluginConfig)
	notificationPipeline.Plugins["notification"] = PluginConfig{
		Order: []string{"com.utmstack.stats"},
	}

	utmstackPipeline := PluginsConfig{}
	utmstackPipeline.Plugins = make(map[string]PluginConfig)
	utmstackPipeline.Plugins["com.utmstack"] = PluginConfig{
		RulesFolder:   "/workdir/rules",
		GeoIPFolder:   "/workdir/geolocation",
		Elasticsearch: "http://node1:9200",
		PostgreSQL: PostgreConfig{
			Server:   "postgres",
			Port:     "5432",
			User:     "postgres",
			Password: conf.Password,
			Database: "utmstack",
		},
		ServerName:   conf.ServerName,
		InternalKey:  conf.InternalKey,
		AgentManager: "10.21.199.3:9000",
		Backend:      "http://backend:8080",
		CertsFolder:  "/cert",
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

	err = utils.WriteYAML(filepath.Join(pipelineDir, "system_plugins_input.yaml"), inputPipeline)
	if err != nil {
		return fmt.Errorf("error writing input pipeline config: %w", err)
	}

	err = utils.WriteYAML(filepath.Join(pipelineDir, "system_plugins_notification.yaml"), notificationPipeline)
	if err != nil {
		return fmt.Errorf("error writing notification pipeline config: %w", err)
	}

	err = utils.WriteYAML(filepath.Join(pipelineDir, "utmstack_plugins.yaml"), utmstackPipeline)
	if err != nil {
		return fmt.Errorf("error writing UTMStack pipeline config: %w", err)
	}

	return nil
}
