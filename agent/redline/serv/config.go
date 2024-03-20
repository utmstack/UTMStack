package serv

import (
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/redline/configuration"
)

// GetConfigServ creates and returns a pointer to a service configuration structure.
func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        configuration.SERV_NAME,
		DisplayName: "UTMStack Redline",
		Description: "UTMStack Redline Service",
	}

	return svcConfig
}
