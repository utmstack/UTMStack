package service

import "github.com/kardianos/service"

func GetConfigServ() *service.Config {
	return &service.Config{
		Name:        "UTMStackEDR",
		DisplayName: "UTMStack EDR",
		Description: "UTMStack Endpoint Detection & Response",
	}
}
