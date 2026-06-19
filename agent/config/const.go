package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/shared/fs"
)

const (
	SERVICE_UPDATER_NAME = "UTMStackUpdater"
)

// DataType identifies the log source type reported to the engine.
type DataType string

const (
	DataTypeLinuxAgent   DataType = "linux"
	DataTypeWindowsAgent DataType = "wineventlog"
	DataTypeMacOs        DataType = "macos"
)

var (
	REPLACE_KEY string

	DependUrl        = "https://%s:%s/private/dependencies/agent/%s"
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"
	DependenciesPort = "9001"

	ServiceLogFile      = filepath.Join(fs.GetExecutablePath(), "logs", "utmstack_agent.log")
	ModulesServName     = "UTMStackModulesLogsCollector"
	UUIDFileName        = filepath.Join(fs.GetExecutablePath(), "uuid.yml")
	ConfigurationFile   = filepath.Join(fs.GetExecutablePath(), "config.yml")
	RetentionConfigFile = filepath.Join(fs.GetExecutablePath(), "retention.json")
	LogsDBFile          = filepath.Join(fs.GetExecutablePath(), "logs_process", "logs.db")
	VersionPath         = filepath.Join(fs.GetExecutablePath(), "version.json")
)
