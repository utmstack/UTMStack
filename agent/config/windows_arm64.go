//go:build windows && arm64
// +build windows,arm64

package config

var (
	UpdaterSelf = "utmstack_updater_self_arm64%s.exe"
	ServiceFile = "utmstack_agent_service_arm64%s.exe"
	DependFiles = []string{"utmstack_agent_dependencies_windows_arm64.zip"}
)
