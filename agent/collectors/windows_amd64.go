//go:build windows && amd64
// +build windows,amd64

package collectors

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/utils"
)

type Windows struct{}

func getCollectorsInstances() []Collector {
	var collectors []Collector
	collectors = append(collectors, Windows{})
	collectors = append(collectors, Filebeat{})

	return collectors
}

func (w Windows) Install() error {
	path := utils.GetMyPath()

	winlogPath := filepath.Join(path, "beats", "winlogbeat")
	beatConfig := CollectorConfig{
		LogsPath:    filepath.Join(winlogPath, "logs"),
		LogFileName: "windowscollector",
	}

	if isInstalled, err := utils.CheckIfServiceIsInstalled(config.WinServName); err != nil {
		return fmt.Errorf("error checking if %s service is installed: %v", config.WinServName, err)
	} else if !isInstalled {
		err = utils.CreatePathIfNotExist(beatConfig.LogsPath)
		if err != nil {
			return fmt.Errorf("error creating %s folder", beatConfig.LogsPath)
		}

		configFile := filepath.Join(winlogPath, "winlogbeat.yml")
		templateFile := filepath.Join(path, "templates", "winlogbeat.yml")
		err = utils.GenerateFromTemplate(beatConfig, templateFile, configFile)
		if err != nil {
			return fmt.Errorf("error configuration from %s: %v", templateFile, err)
		}

		err = utils.Execute("sc",
			winlogPath,
			"create",
			config.WinServName,
			"binPath=",
			fmt.Sprintf("\"%s\\winlogbeat.exe\" --environment=windows_service -c \"%s\\winlogbeat.yml\" --path.home \"%s\" --path.data \"C:\\ProgramData\\winlogbeat\" --path.logs \"C:\\ProgramData\\winlogbeat\\logs\" -E logging.files.redirect_stderr=true", winlogPath, winlogPath, winlogPath),
			"DisplayName=",
			config.WinServName,
			"start=",
			"auto")
		if err != nil {
			return fmt.Errorf("error installing %s service: %s", config.WinServName, err)
		}

		err = utils.Execute("sc", winlogPath, "start", config.WinServName)
		if err != nil {
			return fmt.Errorf("error starting %s service: %s", config.WinServName, err)
		}
	}

	return nil
}

func (w Windows) SendLogs() {
	host, _ := os.Hostname()
	logLinesChan := make(chan string)
	path := utils.GetMyPath()
	winbLogPath := filepath.Join(path, "beats", "winlogbeat", "logs")
	host, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("error getting hostname: %v", err)
		host = "unknown"
	}

	go utils.WatchFolder("windowscollector", winbLogPath, logLinesChan)
	for logLine := range logLinesChan {
		validatedLog, _, err := validations.ValidateString(logLine, false)
		if err != nil {
			utils.Logger.ErrorF("error validating log: %s: %v", logLine, err)
			continue
		}
		logservice.LogQueue <- &plugins.Log{
			DataType:   string(config.DataTypeWindowsAgent),
			DataSource: host,
			Raw:        validatedLog,
		}
	}
}

func (w Windows) Uninstall() error {
	if isInstalled, err := utils.CheckIfServiceIsInstalled(config.WinServName); err != nil {
		return fmt.Errorf("error checking if %s is running: %v", config.WinServName, err)
	} else if isInstalled {
		err = utils.StopService(config.WinServName)
		if err != nil {
			return fmt.Errorf("error stopping %s: %v", config.WinServName, err)
		}
		err = utils.UninstallService(config.WinServName)
		if err != nil {
			return fmt.Errorf("error uninstalling %s: %v", config.WinServName, err)
		}
	}

	return nil
}
