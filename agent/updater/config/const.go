package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/updater/utils"
)

const (
	SERV_LOG        = "utmstack_updater.log"
	SERV_AGENT_NAME = "UTMStackAgent"
)

var (
	DependUrl        = "https://%s:%s/private/dependencies/agent/%s"
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"
	DependenciesPort = "9001"

	VersionPath = filepath.Join(utils.GetMyPath(), "version.json")
)
