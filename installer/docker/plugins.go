package docker

import (
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
	"gopkg.in/yaml.v3"
)

type PluginsConfig struct {
	Plugins map[string]PluginConfig `yaml:"plugins"`
}

type PluginConfig struct {
	RulesFolder   string        `yaml:"rulesFolder"`
	GeoIPFolder   string        `yaml:"geoipFolder"`
	Elasticsearch string        `yaml:"elasticsearch"`
	PostgreSQL    PostgreConfig `yaml:"postgresql"`
	ServerName    string        `yaml:"serverName"`
	InternalKey   string        `yaml:"internalKey"`
	AgentManager  string        `yaml:"agentManager"`
	Backend       string        `yaml:"backend"`
	CertsFolder   string        `yaml:"certsFolder"`
	BdgzPort      string        `yaml:"bdgzPort"`
}

type PostgreConfig struct {
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (c *PluginsConfig) Set(conf *config.Config, stack *StackConfig) error {
	c.Plugins = make(map[string]PluginConfig)

	c.Plugins["com.utmstack"] = PluginConfig{
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

	config, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	pipelineDir := filepath.Join(stack.EventsEngineWorkdir, "pipeline")

	utils.CreatePathIfNotExist(pipelineDir)

	err = os.WriteFile(filepath.Join(pipelineDir, "utmstack_plugins.yaml"), config, 0644)
	if err != nil {
		return err
	}

	return nil
}
