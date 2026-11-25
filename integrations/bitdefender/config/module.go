package config

import (
	"strings"
)

type BDGZModuleConfig struct {
	ConnectionKey string
	AccessUrl     string
	MasterIp      string
	CompaniesIDs  []string
}

func GetBDGZModuleConfig(group *ModuleGroup) BDGZModuleConfig {
	bdgzPro := BDGZModuleConfig{}
	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "connectionKey":
			bdgzPro.ConnectionKey = cnf.ConfValue
		case "accessUrl":
			bdgzPro.AccessUrl = cnf.ConfValue
		case "utmPublicIP":
			bdgzPro.MasterIp = cnf.ConfValue
		case "companyIds":
			bdgzPro.CompaniesIDs = strings.Split(cnf.ConfValue, ",")
		}
	}
	return bdgzPro
}
