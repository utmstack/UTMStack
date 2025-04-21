//go:build darwin && arm64
// +build darwin,arm64

package config

var (
	UpdaterSelf = "utmstack_updater_self%s"
	ServiceFile = "utmstack_agent_service%s"
	DependFiles = []string{}
)
