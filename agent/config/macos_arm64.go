//go:build darwin && arm64
// +build darwin,arm64

package config

var (
	UpdaterSelf = "utmstack_updater_self_arm64%s"
	ServiceFile = "utmstack_agent_service_arm64%s"
	DependFiles = []string{}
)
