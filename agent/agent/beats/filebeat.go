package beats

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/parser"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type Filebeat struct{}

func (f Filebeat) Install() error {
	path := utils.GetMyPath()

	filebLogPath := filepath.Join(path, "beats", "filebeat")
	beatConfig := BeatConfig{
		LogsPath:    filepath.Join(filebLogPath, "logs"),
		LogFileName: "modulescollector",
	}

	if isInstalled, err := utils.CheckIfServiceIsInstalled(config.ModulesServName); err != nil {
		return fmt.Errorf("error checking if %s service is installed: %v", config.ModulesServName, err)
	} else if !isInstalled {
		if err = utils.CreatePathIfNotExist(beatConfig.LogsPath); err != nil {
			return fmt.Errorf("error creating %s folder", beatConfig.LogsPath)
		}

		configFile := filepath.Join(filebLogPath, "filebeat.yml")
		templateFile := filepath.Join(path, "templates", "filebeat.yml")
		if err = utils.GenerateFromTemplate(beatConfig, templateFile, configFile); err != nil {
			return fmt.Errorf("error config from %s: %v", templateFile, err)
		}
		switch runtime.GOOS {
		case "windows":
			err = utils.Execute("sc",
				filebLogPath,
				"create",
				config.ModulesServName,
				"binPath=",
				fmt.Sprintf("\"%s\\filebeat.exe\" --environment=windows_service -c \"%s\\filebeat.yml\" --path.home \"%s\" --path.data \"C:\\ProgramData\\filebeat\" --path.logs \"C:\\ProgramData\\filebeat\\logs\" -E logging.files.redirect_stderr=true", filebLogPath, filebLogPath, filebLogPath),
				"DisplayName=",
				config.ModulesServName,
				"start=",
				"auto")
			if err != nil {
				return fmt.Errorf("error installing %s service: %s", config.ModulesServName, err)
			}

			err = utils.Execute("sc", filebLogPath, "start", config.ModulesServName)
			if err != nil {
				return fmt.Errorf("error starting %s service: %s", config.ModulesServName, err)
			}
		case "linux":
			if err = utils.CreateLinuxService(config.ModulesServName, fmt.Sprintf(
				"%s -c %s -path.home %s -path.config %s -path.data /var/lib/filebeat -path.logs /var/log/filebeat",
				filepath.Join(filebLogPath, "filebeat"),
				filepath.Join(filebLogPath, "filebeat.yml"),
				filebLogPath,
				filebLogPath,
			),
			); err != nil {
				return fmt.Errorf("error creating %s service: %v", config.ModulesServName, err)
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
				err := utils.Execute("systemctl", filebLogPath, "enable", config.ModulesServName)
				if err != nil {
					return fmt.Errorf("%s", err)
				}

				err = utils.Execute("systemctl", filebLogPath, "start", config.ModulesServName)
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

				err = utils.Execute("systemctl", filebLogPath, "restart", config.ModulesServName)
				if err != nil {
					return fmt.Errorf("%s", err)
				}
			}
		}
	}

	return nil
}

func (f Filebeat) SendSystemLogs() {
	host, _ := os.Hostname()
	logLinesChan := make(chan string)
	path := utils.GetMyPath()
	filebLogPath := filepath.Join(path, "beats", "filebeat", "logs")

	parser := parser.GetParser("beats")

	go utils.WatchFolder("modulescollector", filebLogPath, logLinesChan)
	for logLine := range logLinesChan {
		message, _, err := validations.ValidateString(logLine, false)
		if err != nil {
			utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
			continue
		}
		err = parser.ProcessData(logLine, host, logservice.LogQueue)
		if err != nil {
			utils.Logger.ErrorF("error processing beats data: %v", err)
			continue
		}
	}
}

func (f Filebeat) Uninstall() error {
	if isInstalled, err := utils.CheckIfServiceIsInstalled(config.ModulesServName); err != nil {
		return fmt.Errorf("error checking if %s is running: %v", config.ModulesServName, err)
	} else if isInstalled {
		err = utils.StopService(config.ModulesServName)
		if err != nil {
			return fmt.Errorf("error stopping %s: %v", config.ModulesServName, err)
		}

		err = utils.UninstallService(config.ModulesServName)
		if err != nil {
			return fmt.Errorf("error uninstalling %s: %v", config.ModulesServName, err)
		}
	}
	return nil
}
