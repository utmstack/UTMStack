package service

import (
	"github.com/kardianos/service"
)

// GetConfigServ creates and returns a pointer to a service configuration structure.
func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        "UTMStackUpdater",
		DisplayName: "UTMStack Updater",
		Description: "UTMStack Agent Updater Service",
		// No arguments needed - service will just run
	}

	return svcConfig
}
