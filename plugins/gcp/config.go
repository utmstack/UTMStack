package main

import (
	"github.com/utmstack/config-client-go/types"
)

const DefaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

type PluginConfig struct {
	InternalKey string `yaml:"internal_key"`
	Backend     string `yaml:"backend"`
}

type ModuleConfig struct {
	JsonKey        string
	ProjectID      string
	SubscriptionID string
	Topic          string
}

func compareConfigs(saveConfs map[string]ModuleConfig, newConfs []types.ModuleGroup) (isNecessaryConfig bool) {
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
