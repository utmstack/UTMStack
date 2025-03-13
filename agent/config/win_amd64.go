//go:build windows && amd64
// +build windows,amd64

package config

var (
	UpdaterSelf = "utmstack_updater_self%s.exe"
	ServiceFile = "utmstack_agent_service%s.exe"
	DependFiles = []string{"utmstack_agent_dependencies_windows.zip"}
)
