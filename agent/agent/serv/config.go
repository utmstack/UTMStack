package serv

import (
	"github.com/kardianos/service"
)

// GetConfigServ creates and returns a pointer to a service configuration structure.
func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        "UTMStackAgent",
		DisplayName: "UTMStack Agent",
		Description: "UTMStack Agent Service",
	}

	return svcConfig
}
