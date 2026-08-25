package config

import (
	"fmt"
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
	AgentManagerPort = "443"
	LogAuthProxyPort = "443"
	DependenciesPort = "443"

	VersionPath = filepath.Join(fs.GetExecutablePath(), "version.json")
)

// ServiceFile returns the agent binary name with OS and architecture suffix.
// Format: utmstack_agent_service_<os>_<arch>[.exe]
// Examples:
//   - utmstack_agent_service_linux_amd64
//   - utmstack_agent_service_windows_amd64.exe
//   - utmstack_agent_service_darwin_arm64
func ServiceFile(suffix string) string {
	name := fmt.Sprintf("%s_%s_%s%s", agentBaseName, runtime.GOOS, runtime.GOARCH, suffix)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
