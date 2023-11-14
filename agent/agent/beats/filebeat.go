package beats

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/filters"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type Filebeat struct{}

func (f Filebeat) Install(h *holmes.Logger) error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("error getting current path: %v", err)
	}

	filebLogPath := filepath.Join(path, "beats", "filebeat")
	beatConfig := BeatConfig{
		LogsPath:    filepath.Join(filebLogPath, "logs"),
		LogFileName: "modulescollector",
	}

	if isInstalled, err := utils.CheckIfServiceIsInstalled(configuration.ModulesServName); err != nil {
		return fmt.Errorf("error checking if %s service is installed: %v", configuration.ModulesServName, err)
	} else if !isInstalled {
		if err = utils.CreatePathIfNotExist(beatConfig.LogsPath); err != nil {
			return fmt.Errorf("error creating %s folder", beatConfig.LogsPath)
		}

		configFile := filepath.Join(filebLogPath, "filebeat.yml")
		templateFile := filepath.Join(path, "templates", "filebeat.yml")
		if err = utils.GenerateFromTemplate(beatConfig, templateFile, configFile); err != nil {
			return fmt.Errorf("error configuration from %s: %v", templateFile, err)
		}
		switch runtime.GOOS {
		case "windows":
			err = utils.Execute("sc",
				filebLogPath,
				"create",
				configuration.ModulesServName,
				"binPath=",
				fmt.Sprintf("\"%s\\filebeat.exe\" --environment=windows_service -c \"%s\\filebeat.yml\" --path.home \"%s\" --path.data \"C:\\ProgramData\\filebeat\" --path.logs \"C:\\ProgramData\\filebeat\\logs\" -E logging.files.redirect_stderr=true", filebLogPath, filebLogPath, filebLogPath),
				"DisplayName=",
				configuration.ModulesServName,
				"start=",
				"auto")
			if err != nil {
				return fmt.Errorf("error installing %s service: %s", configuration.ModulesServName, err)
			}

			err = utils.Execute("sc", filebLogPath, "start", configuration.ModulesServName)
			if err != nil {
				return fmt.Errorf("error starting %s service: %s", configuration.ModulesServName, err)
			}
		case "linux":
			if err = utils.CreateLinuxService(configuration.ModulesServName, fmt.Sprintf(
				"%s -c %s -path.home %s -path.config %s -path.data /var/lib/filebeat -path.logs /var/log/filebeat",
				filepath.Join(filebLogPath, "filebeat"),
				filepath.Join(filebLogPath, "filebeat.yml"),
				filebLogPath,
				filebLogPath,
			),
			); err != nil {
				return fmt.Errorf("error creating %s service: %v", configuration.ModulesServName, err)
			}

			if err = utils.Execute("chmod", filebLogPath, "-R", "777", "filebeat"); err != nil {
				return fmt.Errorf("error executing chmod: %v", err)
			}

			if err = utils.Execute("systemctl", filebLogPath, "daemon-reload"); err != nil {
				return fmt.Errorf("error reloading daemon: %v", err)
			}

			family, err := utils.DetectLinuxFamily()
			if err != nil {
				return err
			}

			if family == "debian" || family == "rhel" {
				err := utils.Execute("systemctl", filebLogPath, "enable", configuration.ModulesServName)
				if err != nil {
					return fmt.Errorf("%s", err)
				}

				err = utils.Execute("systemctl", filebLogPath, "start", configuration.ModulesServName)
				if err != nil {
					return fmt.Errorf("%s", err)
				}

				err = utils.Execute("./filebeat", filebLogPath, "modules", "enable", "system")
				if err != nil {
					return fmt.Errorf("%s", err)
				}

				err = utils.Execute("sed", filepath.Join(filebLogPath, "modules.d"), "-i", "s/enabled: false/enabled: true/g", "system.yml")
				if err != nil {
					return fmt.Errorf("%s", err)
				}

				err = utils.Execute("systemctl", filebLogPath, "restart", configuration.ModulesServName)
				if err != nil {
					return fmt.Errorf("%s", err)
				}
			}
		}
	}

	return nil
}

func (f Filebeat) SendSystemLogs(h *holmes.Logger) {
	logLinesChan := make(chan []string)
	path, err := utils.GetMyPath()
	if err != nil {
		h.Error("error getting current path: %v", err)
	}
	filebLogPath := filepath.Join(path, "beats", "filebeat", "logs")

	go utils.WatchFolder("modulescollector", filebLogPath, logLinesChan, configuration.BatchCapacity, h)
	for logLine := range logLinesChan {
		beatsData, err := filters.ProcessBeatData(logLine)
		if err != nil {
			h.Error("error processing beats data: %v", err)
			continue
		}
		for typ, logB := range beatsData {
			logservice.LogQueue <- logservice.LogPipe{
				Src:  typ,
				Logs: logB,
			}
		}
	}
}

func (f Filebeat) Uninstall() error {
	if isInstalled, err := utils.CheckIfServiceIsInstalled(configuration.ModulesServName); err != nil {
		return fmt.Errorf("error checking if %s is running: %v", configuration.ModulesServName, err)
	} else if isInstalled {
		err = utils.StopService(configuration.ModulesServName)
		if err != nil {
			return fmt.Errorf("error stopping %s: %v", configuration.ModulesServName, err)
		}

		err = utils.UninstallService(configuration.ModulesServName)
		if err != nil {
			return fmt.Errorf("error uninstalling %s: %v", configuration.ModulesServName, err)
		}
	}
	return nil
}
