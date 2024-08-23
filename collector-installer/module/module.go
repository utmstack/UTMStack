package module

import (
	"github.com/utmstack/UTMStack/collector-installer/config"
)

type ProcessConfig struct {
	ServiceInfo config.ServiceConfig
}

type CollectorProcess interface {
	Run() error
	Install(ip, utmKey, skip string) error
	CheckUpdates()
	Uninstall() error
}

func GetCollectorProcess(collector config.Collector) CollectorProcess {
	switch collector {
	case config.AS400:
		return getAS400Collector()
	default:
		return nil
	}
}
