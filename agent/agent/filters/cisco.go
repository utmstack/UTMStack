package filters

import (
	"regexp"

	"github.com/utmstack/UTMStack/agent/agent/config"
)

var RegexspCisco = map[config.LogType]string{
	config.LogTypeCiscoAsa:       `%ASA-`,
	config.LogTypeCiscoFirepower: `%FTD-`,
	config.LogTypeCiscoSwitch:    `%(\w|_)+-((\b\w+\b-\b\w+\b-)?)(\d)-([A-Z]|_)+`,
}

func indentifyCiscoLogSource(log string) (string, error) {
	for logType, regp := range RegexspCisco {
		regExpCompiled, err := regexp.Compile(string(regp))
		if err != nil {
			return "", err
		}
		if regExpCompiled.MatchString(log) {
			return logType.DataType, nil
		}
	}
	return config.LogTypeCiscoMeraki.DataType, nil
}

func ProcessCiscoData(logBatch []string) (map[string][]string, error) {
	classifiedLogs := make(map[string][]string)
	for _, log := range logBatch {
		if logType, err := indentifyCiscoLogSource(log); err != nil {
			return nil, err
		} else {
			classifiedLogs[string(logType)] = append(classifiedLogs[string(logType)], log)
		}
	}
	return classifiedLogs, nil
}
