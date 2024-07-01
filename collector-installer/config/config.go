package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/collector-installer/utils"
)

type ServiceTypeConfig struct {
	CollectorType Collector `yaml:"collector_type"`
}

const (
	AgentManagerPort    = "9000"
	LogAuthProxyPort    = "50051"
	CollectorConfigFile = "collector-config.yaml"
	SERV_LOG            = "utmstack_collector.log"
)

type ServiceConfig struct {
	Name        string
	DisplayName string
	Description string
	CMDRun      string
	CMDArgs     []string
	CMDPath     string
}

type Collector string

var (
	AS400 Collector = "as400"
)

func IsValidCollector(c string) bool {
	switch c {
	case string(AS400):
		return true
	default:
		return false
	}
}

func SaveConfig(config *ServiceTypeConfig) error {
	path := utils.GetMyPath()
	if err := utils.WriteYAML(filepath.Join(path, CollectorConfigFile), config); err != nil {
		return err
	}
	return nil
}

func ReadConfig() (*ServiceTypeConfig, error) {
	path := utils.GetMyPath()
	config := &ServiceTypeConfig{}
	if err := utils.ReadYAML(filepath.Join(path, CollectorConfigFile), config); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}
	return config, nil
}

func GetDownloadFilePath(typ string, module Collector) string {
	path := utils.GetMyPath()
	switch module {
	case AS400:
		switch typ {
		case "service":
			return filepath.Join(path, "utmstack_collector_as400_service.jar")
		case "dependencies":
			return filepath.Join(path, "dependencies.zip")
		}
	}
	return ""
}

func GetVersionPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "version.json")
}

func CheckIfNecessaryDownloadDependencies(collectorType Collector) bool {
	switch collectorType {
	case AS400:
		switch runtime.GOOS {
		case "windows":
			return true
		case "linux":
			return false
		}
	}
	return false
}
