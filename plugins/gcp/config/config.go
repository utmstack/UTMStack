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

		if cnf.JsonKey != newConf.Configurations[0].ConfValue ||
			cnf.ProjectID != newConf.Configurations[1].ConfValue ||
			cnf.SubscriptionID != newConf.Configurations[2].ConfValue ||
			cnf.Topic != newConf.Configurations[3].ConfValue {

			// utils.Logger.Info("Configuration has been changed")
			// utils.Logger.Info("Old configuration: SubscriptionID-%s, ProjectID-%s, Topic-%s, JsonKey-%s", cnf.SubscriptionID, cnf.ProjectID, cnf.Topic, cnf.JsonKey)
			// utils.Logger.Info("New configuration: SubscriptionID-%s, ProjectID-%s, Topic-%s, JsonKey-%s", newConf.Configurations[2].ConfValue, newConf.Configurations[1].ConfValue, newConf.Configurations[3].ConfValue, newConf.Configurations[0].ConfValue)

			return true
		}
	}
	return false
}
