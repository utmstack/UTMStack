package main

import (
	"fmt"
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

func (c *Config) Get() error {
	config, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(config, c)
	if err != nil {
		return err
	}

	if c.DataDir == "" {
		return fmt.Errorf("data_dir is empty")
	}

	if c.Password == "" {
		return fmt.Errorf("password is empty")
	}

	if c.Branch == "" {
		return fmt.Errorf("branch is empty")
	}

	if c.MainServer == "" {
		return fmt.Errorf("main_server is empty")
	}

	if c.ServerType == "" {
		return fmt.Errorf("server_type is empty")
	}

	if c.ServerName == "" {
		return fmt.Errorf("server_name is empty")
	}

	if c.InternalKey == "" {
		return fmt.Errorf("internal_key is empty")
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
