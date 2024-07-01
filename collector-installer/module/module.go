package module

import (
	"github.com/utmstack/UTMStack/collector-installer/config"
	"github.com/utmstack/UTMStack/collector-installer/utils"
)

type ProcessConfig struct {
	ServiceInfo config.ServiceConfig
	Logger      *utils.BeautyLogger
}

type CollectorProcess interface {
	Run() error
	Install(ip string, utmKey string) error
	Uninstall() error
}

func GetCollectorProcess(collector config.Collector, logger *utils.BeautyLogger) CollectorProcess {
	switch collector {
	case config.AS400:
		return getAS400Collector(logger)
	default:
		return nil
	}
}
