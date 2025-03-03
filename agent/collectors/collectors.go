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
		collectors = append(collectors, Winlogbeat{})
		collectors = append(collectors, Filebeat{})
	case "linux":
		collectors = append(collectors, Filebeat{})
	case "macos":
		collectors = append(collectors, MacOS{})
	}

	return collectors
}

func InstallCollectors() error {
	collectors := getCollectorsInstances()

	for _, beat := range collectors {
		err := beat.Install()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	utils.Logger.Info("collectors installed correctly")

	return nil
}

func LogsReader() {
	collectors := getCollectorsInstances()
	for _, beat := range collectors {
		go beat.SendSystemLogs()
	}
}

func UninstallCollectors() error {
	collectors := getCollectorsInstances()

	for _, beat := range collectors {
		err := beat.Uninstall()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	utils.Logger.Info("collectors uninstalled correctly")
	return nil
}
