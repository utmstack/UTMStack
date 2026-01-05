//go:build windows && amd64
// +build windows,amd64

package config

var (
	ServiceFile = "utmstack_agent_service%s.exe"
	UpdaterFile = "utmstack_updater_service%s"
	DependFiles = []string{"utmstack_agent_dependencies_windows.zip"}
)
