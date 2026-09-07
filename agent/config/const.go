package config

import (
	"os"
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
	AgentManagerPort = "443"
	LogAuthProxyPort = "443"
	DependenciesPort = "443"

	ServiceLogFile      = filepath.Join(fs.GetExecutablePath(), "logs", "utmstack_agent.log")
	ModulesServName     = "UTMStackModulesLogsCollector"
	UUIDFileName        = filepath.Join(fs.GetExecutablePath(), "uuid.yml")
	ConfigurationFile   = filepath.Join(fs.GetExecutablePath(), "config.yml")
	RetentionConfigFile = filepath.Join(fs.GetExecutablePath(), "retention.json")
	ConfigStateFile     = filepath.Join(fs.GetExecutablePath(), "config_state.json")
	UpdateHoldFile      = filepath.Join(fs.GetExecutablePath(), "update_hold.json")
	LogsDBFile          = filepath.Join(fs.GetExecutablePath(), "logs_process", "logs.db")
	VersionPath         = filepath.Join(fs.GetExecutablePath(), "version.json")
)

// In production all three ports are 443, fronted by a single reverse proxy
// that routes by path/ALPN to agent-manager and log-input. Overridable so
// the agent can talk to those services directly — e.g. a local dev stack
// with no proxy in front, where agent-manager's gRPC and dependencies
// server and log-input's gRPC are each on their own plain port.
func init() {
	if v := os.Getenv("UTMSTACK_AGENT_MANAGER_PORT"); v != "" {
		AgentManagerPort = v
	}
	if v := os.Getenv("UTMSTACK_LOG_AUTH_PROXY_PORT"); v != "" {
		LogAuthProxyPort = v
	}
	if v := os.Getenv("UTMSTACK_DEPENDENCIES_PORT"); v != "" {
		DependenciesPort = v
	}
}
