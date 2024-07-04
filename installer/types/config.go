package types

import (
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MainServer  string `yaml:"main_server"`
	Branch      string `yaml:"branch"`
	Password    string `yaml:"password"`
	DataDir     string `yaml:"data_dir"`
	ServerType  string `yaml:"server_type"`
	ServerName  string `yaml:"server_name"`
	InternalKey string `yaml:"internal_key"`
}

type ModuleConfig struct {
	Branch string `yaml:"branch"`
}

var configPath = filepath.Join("/root", "utmstack.yml")

func (c *Config) Get() {
	config, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	_ = yaml.Unmarshal(config, c)
}

func (c *Config) Set() error {
	config, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, config, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) SetModulesConfig() error {
	mconfig, err := yaml.Marshal(&ModuleConfig{Branch: c.Branch})
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(c.DataDir, "agent_manager", "config.yml"), mconfig, 0644)
	if err != nil {
		return err
	}

	return nil
}
