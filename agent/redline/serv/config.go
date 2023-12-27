package serv

import (
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/redline/constants"
)

// GetConfigServ creates and returns a pointer to a service configuration structure.
func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        constants.SERV_NAME,
		DisplayName: "UTMStack Redline",
		Description: "UTMStack Redline Service",
	}

	return svcConfig
}
