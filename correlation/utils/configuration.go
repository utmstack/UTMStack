package utils

import (
	"sync"
)

type YamlConfig struct {
	RulesFolder   string `yaml:"rulesFolder"`
	Elasticsearch string `yaml:"elasticsearch"`
	Postgres      struct {
		Server   string `yaml:"server"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"postgresql"`
	ErrorLevel string `yaml:"errorLevel"`
}

type ConnMode string

const (
	ConnModeOffline ConnMode = "OFFLINE"
	ConnModeOnline  ConnMode = "ONLINE"
)

type EnvVarConfig struct {
	ConnectionMode ConnMode `env:"CONNECTION_MODE"`
}

type Config struct {
	YamlConfig
	EnvVarConfig
}

var oneConfigRead sync.Once
var cnf Config

func readConfig() {
	ReadYaml("config.yml", &cnf.YamlConfig)
	ReadEnvVars(&cnf.EnvVarConfig)
}

func GetConfig() Config {
	oneConfigRead.Do(func() { readConfig() })
	return cnf
}
