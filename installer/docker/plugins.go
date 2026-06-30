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
	Order         []string         `yaml:"order,omitempty"`
	Port          int              `yaml:"port,omitempty"`
	RulesFolder   string           `yaml:"rulesFolder,omitempty"`
	GeoIPFolder   string           `yaml:"geoipFolder,omitempty"`
	OpenSearch    OpenSearchConfig `yaml:"opensearch,omitempty"`
	PostgreSQL    PostgreConfig    `yaml:"postgresql,omitempty"`
	ServerName    string           `yaml:"serverName,omitempty"`
	InternalKey   string           `yaml:"internalKey,omitempty"`
	EncryptionKey string           `yaml:"encryptionKey,omitempty"`
	Env           string           `yaml:"env,omitempty"`
	AgentManager  string           `yaml:"agentManager,omitempty"`
	Backend       string           `yaml:"backend,omitempty"`
	CertsFolder   string           `yaml:"certsFolder,omitempty"`
	ModulesConfig string           `yaml:"modulesConfig,omitempty"`
}

type PostgreConfig struct {
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type OpenSearchConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func SetPluginsConfigs(conf *config.Config, stack *StackConfig) error {
	analysisPipeline := PluginsConfig{}
	analysisPipeline.Plugins = make(map[string]PluginConfig)
	analysisPipeline.Plugins["analysis"] = PluginConfig{
		Order: []string{"com.utmstack.events", "cel", "feeds"},
	}

	correlationPipeline := PluginsConfig{}
	correlationPipeline.Plugins = make(map[string]PluginConfig)
	correlationPipeline.Plugins["correlation"] = PluginConfig{
		Order: []string{"com.utmstack.alerts", "com.utmstack.soc-ai"},
	}

	notificationPipeline := PluginsConfig{}
	notificationPipeline.Plugins = make(map[string]PluginConfig)
	notificationPipeline.Plugins["notification"] = PluginConfig{
		Order: []string{"com.utmstack.stats"},
	}

	utmstackPipeline := PluginsConfig{}
	utmstackPipeline.Plugins = make(map[string]PluginConfig)
	utmstackPipeline.Plugins["com.utmstack"] = PluginConfig{
		RulesFolder: "/workdir/rules",
		GeoIPFolder: "/workdir/geolocation",
		OpenSearch: OpenSearchConfig{
			Host:     "node1",
			Port:     "9200",
			User:     "admin",
			Password: conf.OpenSearchPassword,
		},
		PostgreSQL: PostgreConfig{
			Server:   "postgres",
			Port:     "5432",
			User:     "postgres",
			Password: conf.Password,
			Database: "utmstack",
		},
		ServerName:    conf.ServerName,
		InternalKey:   conf.InternalKey,
		EncryptionKey: conf.InternalKey,
		Env:           conf.Branch,
		AgentManager:  "10.21.199.3:9000",
		Backend:       "http://backend:8080",
		CertsFolder:   "/cert",
		ModulesConfig: "event-processor-manager:9003",
	}

	openSearchPipeline := PluginsConfig{}
	openSearchPipeline.Plugins = make(map[string]PluginConfig)
	openSearchPipeline.Plugins["org.opensearch"] = PluginConfig{
		OpenSearch: OpenSearchConfig{
			Host:     "node1",
			Port:     "9200",
			User:     "admin",
			Password: conf.OpenSearchPassword,
		},
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

	err = utils.WriteYAML(filepath.Join(pipelineDir, "opensearch_plugins.yaml"), openSearchPipeline)
	if err != nil {
		return fmt.Errorf("error writing OpenSearch pipeline config: %w", err)
	}

	return nil
}
