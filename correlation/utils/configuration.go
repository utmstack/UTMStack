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

type Config struct {
	YamlConfig
}

var oneConfigRead sync.Once
var cnf Config

func readConfig() {
	ReadYaml("config.yml", &cnf.YamlConfig)
}

func GetConfig() Config {
	oneConfigRead.Do(func() { readConfig() })
	return cnf
}
