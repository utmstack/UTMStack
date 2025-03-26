package collectors

import (
	"os"
	"path/filepath"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/utils"
)

func SendSystemLogs() {
	logLinesChan := make(chan string)
	path := utils.GetMyPath()
	winlogbeatLogPath := filepath.Join(path, "beats", "winlogbeat", "logs")

	go utils.WatchFolder("windowscollector", winlogbeatLogPath, logLinesChan)
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
		logservice.LogQueue <- &plugins.Log{
			DataType:   string(config.DataTypeWindowsAgent),
			DataSource: host,
			Raw:        validatedLog,
		}
	}
}
