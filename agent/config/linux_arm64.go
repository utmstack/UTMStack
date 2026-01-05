//go:build linux && arm64
// +build linux,arm64

package config

var (
	ServiceFile = "utmstack_agent_service_arm64%s"
	UpdaterFile = "utmstack_updater_service%s"
	DependFiles = []string{"utmstack_agent_dependencies_linux_arm64.zip"}
)
