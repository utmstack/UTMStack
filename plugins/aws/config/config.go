package config

import (
	"github.com/utmstack/UTMStack/plugins/aws/schema"
	"github.com/utmstack/config-client-go/types"
)

func CompareConfigs(savedConfs map[string]schema.AWSConfig, newConfs []types.ModuleGroup) bool {
	for _, newConf := range newConfs {
		cnf, ok := savedConfs[newConf.GroupName]
		if !ok {
			return true
		}

		if cnf.AccessKeyID != newConf.Configurations[0].ConfValue ||
			cnf.SecretAccessKey != newConf.Configurations[1].ConfValue ||
			cnf.Region != newConf.Configurations[2].ConfValue ||
			cnf.LogGroupName != newConf.Configurations[3].ConfValue ||
			cnf.LogStreamName != newConf.Configurations[4].ConfValue {
			return true
		}
	}
	return false
}
