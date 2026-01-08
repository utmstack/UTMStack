//go:build darwin
// +build darwin

package config

var (
	ServiceFile = "utmstack_agent_service%s"
	UpdaterFile = "utmstack_updater_service%s"
	DependFiles = []string{}
)
