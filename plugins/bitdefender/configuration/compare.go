package configuration

import (
	"strings"

	"github.com/utmstack/config-client-go/types"
)

func compareConfigs(saveConfs map[string]ModuleConfig, newConf types.ModuleGroup) (isNecessaryConfig bool) {
	cnf, ok := saveConfs[newConf.GroupName]
	if !ok {
		return true
	}

	return isDifferent(cnf.CompaniesIDs, strings.Split(newConf.Configurations[3].ConfValue, ",")) ||
		cnf.ConnectionKey != newConf.Configurations[0].ConfValue ||
		cnf.AccessUrl != newConf.Configurations[1].ConfValue ||
		cnf.MasterIp != newConf.Configurations[2].ConfValue
}

func isDifferent(a1 []string, a2 []string) bool {
	m := make(map[string]bool)

	for _, v := range a1 {
		m[v] = true
	}

	for _, v := range a2 {
		if !m[v] {
			return true
		}
		delete(m, v)
	}

	return len(m) > 0
}
