package service

import (
	"github.com/kardianos/service"
)

func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        "UTMStackAS400Updater",
		DisplayName: "UTMStack AS400 Updater",
		Description: "UTMStack AS400 Collector Updater Service",
	}

	return svcConfig
}
