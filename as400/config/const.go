package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/as400/utils"
)

var REPLACE_KEY string = ""

const (
	DataType string = "ibm-as400"
)

var (
	DependUrl        = "https://%s:%s/private/dependencies/collector/as400/%s"
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"
	DependenciesPort = "9001"

	ServiceLogFile      = filepath.Join(utils.GetMyPath(), "logs", "utmstack_as400_collector.log")
	UUIDFileName        = filepath.Join(utils.GetMyPath(), "uuid.yml")
	ConfigurationFile   = filepath.Join(utils.GetMyPath(), "config.yml")
	RetentionConfigFile = filepath.Join(utils.GetMyPath(), "retention.json")
	LogsDBFile          = filepath.Join(utils.GetMyPath(), "logs_process", "logs.db")
	VersionPath         = filepath.Join(utils.GetMyPath(), "version.json")
	CollectorJarPath    = filepath.Join(utils.GetMyPath(), "as400-collector.jar")
	ConfigFilePath      = filepath.Join(utils.GetMyPath(), "local_storage", "server.json")
)
