package parser

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/utmstack/UTMStack/agent/agent/config"
)

var (
	ciscoParser     = CiscoParser{}
	ciscoParserOnce sync.Once
	RegexspCisco    = map[config.LogType]string{
		config.LogTypeCiscoAsa:       `%ASA-`,
		config.LogTypeCiscoFirepower: `%FTD-`,
		config.LogTypeCiscoSwitch:    `%(\w|_)+-((\b\w+\b-\b\w+\b-)?)(\d)-([A-Z]|_)+`,
	}
)

type CiscoParser struct{}

func GetCiscoParser() *CiscoParser {
	ciscoParserOnce.Do(func() {
		ciscoParser = CiscoParser{}
	})
	return &ciscoParser
}

func (p *CiscoParser) IdentifySource(log string) (config.LogType, error) {
	for logType, regp := range RegexspCisco {
		regExpCompiled, err := regexp.Compile(string(regp))
		if err != nil {
			return "", err
		}
		if regExpCompiled.MatchString(log) {
			return logType, nil
		}
	}
	return config.LogTypeCiscoMeraki, nil
}

func (p *CiscoParser) ProcessData(logBatch interface{}) (map[string][]string, error) {
	classifiedLogs := make(map[string][]string)
	batch, ok := logBatch.([]string)
	if !ok {
		return nil, fmt.Errorf("logBatch is not of type []string")
	}
	for _, log := range batch {
		if logType, err := p.IdentifySource(log); err != nil {
			return nil, err
		} else {
			classifiedLogs[string(logType)] = append(classifiedLogs[string(logType)], log)
		}
	}
	return classifiedLogs, nil
}
