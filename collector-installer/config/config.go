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
	DEPEND_URL          = "https://%s/dependencies/collector?version=%s&os=%s&type=%s&collectorType=%s"
	DependServiceLabel  = "service"
	DependZipLabel      = "depend_zip"
	UpdateLockFile      = "updating.lock"
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

func GetDownloadFilePath(typ string, module Collector, subfix string) string {
	path := utils.GetMyPath()
	switch module {
	case AS400:
		switch typ {
		case DependServiceLabel:
			return filepath.Join(path, fmt.Sprintf("utmstack_collector_as400_service%s.jar", subfix))
		case DependZipLabel:
			switch runtime.GOOS {
			case "windows":
				return filepath.Join(path, "dependencies.zip")
			}
		}
	}
	return ""
}

func GetUpdateLockFilePath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, UpdateLockFile)
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
