package beats

import (
	"fmt"
	"runtime"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
)

type BeatConfig struct {
	LogsPath    string
	LogFileName string
}

type Beat interface {
	Install() error
	SendSystemLogs(h *logger.Logger)
	Uninstall() error
}

func getBeatsInstances() []Beat {
	var beatsInstance []Beat
	switch runtime.GOOS {
	case "windows":
		beatsInstance = append(beatsInstance, Winlogbeat{})
		beatsInstance = append(beatsInstance, Filebeat{})
	case "linux":
		beatsInstance = append(beatsInstance, Filebeat{})
	}

	return beatsInstance
}

func InstallBeats(cnf configuration.Config, h *logger.Logger) error {
	beatsInstances := getBeatsInstances()

	for _, beat := range beatsInstances {
		err := beat.Install()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	h.Info("beats installed correctly")

	return nil
}

func BeatsLogsReader(h *logger.Logger) {
	beatsInstances := getBeatsInstances()
	for _, beat := range beatsInstances {
		go beat.SendSystemLogs(h)
	}
}

func UninstallBeats(h *logger.Logger) error {
	beatsInstances := getBeatsInstances()

	for _, beat := range beatsInstances {
		err := beat.Uninstall()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	h.Info("beats uninstalled correctly")
	return nil
}
