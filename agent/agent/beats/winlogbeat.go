package beats

import (
	"fmt"
	"os"
	"path/filepath"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type Winlogbeat struct{}

func (w Winlogbeat) Install() error {
	path := utils.GetMyPath()

	winlogPath := filepath.Join(path, "beats", "winlogbeat")
	beatConfig := BeatConfig{
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
			return fmt.Errorf("error config from %s: %v", templateFile, err)
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

func (w Winlogbeat) SendSystemLogs() {
	logLinesChan := make(chan string)
	path := utils.GetMyPath()
	winbLogPath := filepath.Join(path, "beats", "winlogbeat", "logs")

	go utils.WatchFolder("windowscollector", winbLogPath, logLinesChan)
	for logLine := range logLinesChan {
		validatedLog, _, err := validations.ValidateString(logLine, false)
		if err != nil {
			utils.Logger.ErrorF("error validating log: %s: %v", logLine, err)
			continue
		}
		host, err := os.Hostname()
		if err != nil {
			utils.Logger.ErrorF("error getting hostname: %v", err)
			host = "unknown"
		}
		logservice.LogQueue <- &go_sdk.Log{
			DataType:   string(config.DataTypeWindowsAgent),
			DataSource: host,
			Raw:        validatedLog,
		}
	}
}

func (w Winlogbeat) Uninstall() error {
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
