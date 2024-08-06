package config

import (
	"github.com/utmstack/UTMStack/plugins/gcp/schema"
	"github.com/utmstack/config-client-go/types"
)

func CompareConfigs(saveConfs map[string]schema.ModuleConfig, newConfs []types.ModuleGroup) (isNecessaryConfig bool) {
	for _, newConf := range newConfs {
		cnf, ok := saveConfs[newConf.GroupName]
		if !ok {
			return true
		}

		if cnf.Topic != newConf.Configurations[3].ConfValue ||
			cnf.SubscriptionID != newConf.Configurations[0].ConfValue ||
			cnf.ProjectID != newConf.Configurations[1].ConfValue ||
			cnf.JsonKey != newConf.Configurations[2].ConfValue {
			return true
		}
	}
	return false
}
