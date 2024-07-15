package parser

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/threatwinds/go-sdk/plugins"
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

func (p *CiscoParser) IdentifySource(log string) (string, error) {
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

func (p *CiscoParser) ProcessData(logMessage interface{}, datasource string, queue chan *plugins.Log) error {
	log, ok := logMessage.(string)
	if !ok {
		return fmt.Errorf("log is not of type string")
	}
	logType, err := p.IdentifySource(log)
	if err != nil {
		return err
	}
	queue <- &plugins.Log{
		DataType:   logType,
		DataSource: datasource,
		Raw:        log,
	}
	return nil
}
