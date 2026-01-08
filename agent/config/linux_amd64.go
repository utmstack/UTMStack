//go:build linux && amd64
// +build linux,amd64

package config

var (
	ServiceFile = "utmstack_agent_service%s"
	UpdaterFile = "utmstack_updater_service%s"
	DependFiles = []string{"utmstack_agent_dependencies_linux.zip"}
)
