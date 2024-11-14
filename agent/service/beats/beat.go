package beats

import (
	"fmt"
	"runtime"

	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/utils"
)

type BeatConfig struct {
	LogsPath    string
	LogFileName string
}

type Beat interface {
	Install() error
	SendSystemLogs()
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

func InstallBeats(cnf config.Config) error {
	beatsInstances := getBeatsInstances()

	for _, beat := range beatsInstances {
		err := beat.Install()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	utils.Logger.Info("beats installed correctly")

	return nil
}

func BeatsLogsReader() {
	beatsInstances := getBeatsInstances()
	for _, beat := range beatsInstances {
		go beat.SendSystemLogs()
	}
}

func UninstallBeats() error {
	beatsInstances := getBeatsInstances()

	for _, beat := range beatsInstances {
		err := beat.Uninstall()
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	utils.Logger.Info("beats uninstalled correctly")
	return nil
}
