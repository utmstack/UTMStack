package serv

import (
	"github.com/kardianos/service"
)

// GetConfigServ creates and returns a pointer to a service configuration structure.
func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        "UTMStackAS400Collector",
		DisplayName: "UTMStack AS400 Collector",
		Description: "UTMStack AS400 Collector Service",
		Arguments:   []string{"run"},
	}

	return svcConfig
}
