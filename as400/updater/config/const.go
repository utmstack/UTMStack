package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/as400/updater/utils"
)

const (
	SERV_LOG            = "utmstack_as400_updater.log"
	SERV_COLLECTOR_NAME = "UTMStackAS400Collector"
	JAR_FILE            = "as400-collector.jar"
)

var (
	DependUrl        = "https://%s:%s/private/dependencies/collector/as400/%s"
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"
	DependenciesPort = "9001"

	ServiceFile = "utmstack_as400_collector_service%s"
	VersionPath = filepath.Join(utils.GetMyPath(), "version.json")
)
