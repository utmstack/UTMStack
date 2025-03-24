package collectors

import (
	"fmt"

	"github.com/utmstack/UTMStack/agent/utils"
)

type CollectorConfig struct {
	LogsPath    string
	LogFileName string
}

type Collector interface {
	Install() error
	SendSystemLogs()
	Uninstall() error
}

func InstallCollectors() error {
	collectors := getCollectorsInstances()

	for _, collector := range collectors {
		err := collector.Install()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	utils.Logger.LogF(100, "collectors installed correctly")

	return nil
}

func LogsReader() {
	collectors := getCollectorsInstances()
	for _, collector := range collectors {
		go collector.SendSystemLogs()
	}
}

func UninstallCollectors() error {
	collectors := getCollectorsInstances()

	for _, collector := range collectors {
		err := collector.Uninstall()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	utils.Logger.LogF(100, "collectors uninstalled correctly")
	return nil
}
