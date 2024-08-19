package config

import (
	"github.com/utmstack/UTMStack/plugins/gcp/schema"
	"github.com/utmstack/config-client-go/types"
)

const DefaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

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

			return true
		}
	}
	return false
}
