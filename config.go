package main

import (
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Branch     string `yaml:"branch"`
	Password   string `yaml:"password"`
	MainServer string `yaml:"main_server"`
	DataDir    string `yaml:"data_dir"`
}

var configPath = path.Join("/root", "utmstack.yml")

func (c *Config) Get() error {
	config, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(config, c)
	if err != nil {
		return err
	}

	return nil
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
