package parser

import (
	"fmt"
	"regexp"
	"sync"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/agent/service/config"
)

var (
	ciscoParser     = CiscoParser{}
	ciscoParserOnce sync.Once
	RegexspCisco    = map[config.DataType]string{
		config.DataTypeCiscoAsa:       `%ASA-`,
		config.DataTypeCiscoFirepower: `%FTD-`,
		config.DataTypeCiscoSwitch:    `%(\w|_)+-((\b\w+\b-\b\w+\b-)?)(\d)-([A-Z]|_)+`,
	}
)

type CiscoParser struct{}

func GetCiscoParser() *CiscoParser {
	ciscoParserOnce.Do(func() {
		ciscoParser = CiscoParser{}
	})
	return &ciscoParser
}

func (p *CiscoParser) IdentifySource(log string) (config.DataType, error) {
	for logType, regp := range RegexspCisco {
		regExpCompiled, err := regexp.Compile(string(regp))
		if err != nil {
			return "", err
		}
		if regExpCompiled.MatchString(log) {
			return logType, nil
		}
	}
	return config.DataTypeCiscoMeraki, nil
}

func (p *CiscoParser) ProcessData(logMessage interface{}, datasource string, queue chan *go_sdk.Log) error {
	log, ok := logMessage.(string)
	if !ok {
		return fmt.Errorf("log is not of type string")
	}
	logType, err := p.IdentifySource(log)
	if err != nil {
		return err
	}
	queue <- &go_sdk.Log{
		DataType:   string(logType),
		DataSource: datasource,
		Raw:        log,
	}
	return nil
}
