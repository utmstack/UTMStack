package main

import (
	"os"
	"path"

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

var configPath = path.Join("/root", "utmstack.yml")

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
