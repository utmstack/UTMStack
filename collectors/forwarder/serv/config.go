package serv

import (
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
)

func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        config.ServiceName,
		DisplayName: "UTMStack Forwarder",
		Description: "UTMStack Forwarder — collects logs from external integrations and forwards them to the backend",
		Arguments:   config.ServiceArguments,
	}

	return svcConfig
}
