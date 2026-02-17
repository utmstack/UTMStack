package config

import (
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/shared/fs"
)

const (
	SERV_LOG        = "utmstack_updater.log"
	SERV_AGENT_NAME = "UTMStackAgent"

	agentBaseName = "utmstack_agent_service"
)

var (
	DependUrl        = "https://%s:%s/private/dependencies/agent/%s"
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"
	DependenciesPort = "9001"

	VersionPath = filepath.Join(fs.GetExecutablePath(), "version.json")
)

// ServiceFile returns the agent binary name with the appropriate suffix and extension.
func ServiceFile(suffix string) string {
	// Check if this is arm64
	isArm64 := runtime.GOARCH == "arm64"

	base := agentBaseName
	if isArm64 {
		base = agentBaseName + "_arm64"
	}

	if runtime.GOOS == "windows" {
		return base + suffix + ".exe"
	}
	return base + suffix
}
