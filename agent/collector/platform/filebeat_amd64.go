package platform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

// beatsRegexps maps data types to their identifying regex patterns.
var beatsRegexps = map[config.DataType]*regexp.Regexp{
	config.DataTypeApacheModule:        regexp.MustCompile(`"type":"apache"|"module":"apache"`),
	config.DataTypeLinuxAuditdModule:   regexp.MustCompile(`"type":"auditd"|"module":"auditd"`),
	config.DataTypeElasticsearchModule: regexp.MustCompile(`"type":"elasticsearch"|"module":"elasticsearch"`),
	config.DataTypeKafkaModule:         regexp.MustCompile(`"type":"kafka"|"module":"kafka"`),
	config.DataTypeKibanaModule:        regexp.MustCompile(`"type":"kibana"|"module":"kibana"`),
	config.DataTypeLogstashModule:      regexp.MustCompile(`"type":"logstash"|"module":"logstash"`),
	config.DataTypeMongodbModule:       regexp.MustCompile(`"type":"mongodb"|"module":"mongodb"`),
	config.DataTypeMysqlModule:         regexp.MustCompile(`"type":"mysql"|"module":"mysql"`),
	config.DataTypeNginxModule:         regexp.MustCompile(`"type":"nginx"|"module":"nginx"`),
	config.DataTypeOsqueryModule:       regexp.MustCompile(`"type":"osquery"|"module":"osquery"`),
	config.DataTypePostgresqlModule:    regexp.MustCompile(`"type":"postgresql"|"module":"postgresql"`),
	config.DataTypeRedisModule:         regexp.MustCompile(`"type":"redis"|"module":"redis"`),
	config.DataTypeLinuxAgent:          regexp.MustCompile(`"type":"system"|"module":"system"`),
	config.DataTypeIisModule:           regexp.MustCompile(`"type":"iis"|"module":"iis"`),
	config.DataTypeTraefikModule:       regexp.MustCompile(`"type":"traefik"|"module":"traefik"`),
	config.DataTypeNatsModule:          regexp.MustCompile(`"type":"nats"|"module":"nats"`),
	config.DataTypeHaproxyModule:       regexp.MustCompile(`"type":"haproxy"|"module":"haproxy"`),
}

// identifyBeatsSource determines the data type of a log message.
func identifyBeatsSource(log string) config.DataType {
	for dataType, re := range beatsRegexps {
		if re.MatchString(log) {
			return dataType
		}
	}
	return config.DataTypeGeneric
}

type Filebeat struct{}

func (f Filebeat) Name() string {
	return "filebeat"
}

func (f Filebeat) Install() error {
	path := fs.GetExecutablePath()

	filebLogPath := filepath.Join(path, "beats", "filebeat")
	beatConfig := CollectorConfig{
		LogsPath:    filepath.Join(filebLogPath, "logs"),
		LogFileName: "modulescollector",
	}

	if isInstalled, err := utils.CheckIfServiceIsInstalled(config.ModulesServName); err != nil {
		return fmt.Errorf("error checking if %s service is installed: %v", config.ModulesServName, err)
	} else if !isInstalled {
		if err = fs.CreateDirIfNotExist(beatConfig.LogsPath); err != nil {
			return fmt.Errorf("error creating %s folder", beatConfig.LogsPath)
		}

		configFile := filepath.Join(filebLogPath, "filebeat.yml")
		templateFile := filepath.Join(path, "templates", "filebeat.yml")
		if err = utils.GenerateFromTemplate(beatConfig, templateFile, configFile); err != nil {
			return fmt.Errorf("error configuration from %s: %v", templateFile, err)
		}
		switch runtime.GOOS {
		case "windows":
			err = exec.Run("sc",
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

			err = exec.Run("sc", filebLogPath, "start", config.ModulesServName)
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

			if err = exec.Run("chmod", filebLogPath, "-R", "755", "filebeat"); err != nil {
				return fmt.Errorf("error executing chmod: %v", err)
			}

			if err = exec.Run("systemctl", filebLogPath, "daemon-reload"); err != nil {
				return fmt.Errorf("error reloading daemon: %v", err)
			}

			err := exec.Run("systemctl", filebLogPath, "enable", config.ModulesServName)
			if err != nil {
				return fmt.Errorf("%s", err)
			}

			err = exec.Run("systemctl", filebLogPath, "start", config.ModulesServName)
			if err != nil {
				return fmt.Errorf("%s", err)
			}

			err = exec.Run("./filebeat", filebLogPath, "modules", "enable", "system")
			if err != nil {
				return fmt.Errorf("%s", err)
			}

			err = exec.Run("sed", filepath.Join(filebLogPath, "modules.d"), "-i", "s/enabled: false/enabled: true/g", "system.yml")
			if err != nil {
				return fmt.Errorf("%s", err)
			}

			err = exec.Run("systemctl", filebLogPath, "restart", config.ModulesServName)
			if err != nil {
				return fmt.Errorf("%s", err)
			}
		}
	}

	return nil
}

func (f Filebeat) Start(ctx context.Context, queue chan *plugins.Log) {
	logLinesChan := make(chan string)
	path := fs.GetExecutablePath()
	filebLogPath := filepath.Join(path, "beats", "filebeat", "logs")
	host, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("error getting hostname: %v", err)
		host = "unknown"
	}

	go utils.WatchFolder("modulescollector", filebLogPath, logLinesChan)

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("filebeat collector stopping due to context cancellation")
			return
		case logLine, ok := <-logLinesChan:
			if !ok {
				return
			}
			message, _, err := entities.ValidateString(logLine, false)
			if err != nil {
				utils.Logger.ErrorF("error validating string: %v: message: %s", err, message)
				continue
			}

			dataType := identifyBeatsSource(logLine)
			queue <- &plugins.Log{
				DataType:   string(dataType),
				DataSource: host,
				Raw:        logLine,
			}
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

func (f Filebeat) Stop() {}
