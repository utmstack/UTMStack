//go:build linux && amd64
// +build linux,amd64

package config

var (
	UpdaterSelf = "utmstack_updater_self%s"
	ServiceFile = "utmstack_agent_service%s"
	DependFiles = []string{"utmstack_agent_dependencies_linux.zip"}
)
