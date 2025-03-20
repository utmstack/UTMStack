package collectors

import (
	"fmt"
	"runtime"

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

func getCollectorsInstances() []Collector {
	var collectors []Collector
	switch runtime.GOOS {
	case "windows":
		collectors = append(collectors, Windows{})
		if runtime.GOARCH == "amd64" {
			collectors = append(collectors, Filebeat{})
		}
	case "linux":
		collectors = append(collectors, Filebeat{})
	}

	return collectors
}

func InstallCollectors() error {
	collectors := getCollectorsInstances()

	for _, collector := range collectors {
		err := collector.Install()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	utils.Logger.Info("collector installed correctly")

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

	utils.Logger.Info("collectors uninstalled correctly")
	return nil
}
