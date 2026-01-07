//go:build linux && arm64
// +build linux,arm64

package config

var (
	UpdaterSelf = "utmstack_updater_self_arm64%s"
	ServiceFile = "utmstack_agent_service_arm64%s"
	DependFiles = []string{"utmstack_agent_dependencies_linux_arm64.zip"}
)
