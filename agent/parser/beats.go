package parser

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

var (
	beatsParser     = BeatsParser{}
	beatsParserOnce sync.Once
	RegexspBeats    = map[config.DataType]string{
		config.DataTypeApacheModule:        `"type":"apache"|"module":"apache"`,
		config.DataTypeLinuxAuditdModule:   `"type":"auditd"|"module":"auditd"`,
		config.DataTypeElasticsearchModule: `"type":"elasticsearch"|"module":"elasticsearch"`,
		config.DataTypeKafkaModule:         `"type":"kafka"|"module":"kafka"`,
		config.DataTypeKibanaModule:        `"type":"kibana"|"module":"kibana"`,
		config.DataTypeLogstashModule:      `"type":"logstash"|"module":"logstash"`,
		config.DataTypeMongodbModule:       `"type":"mongodb"|"module":"mongodb"`,
		config.DataTypeMysqlModule:         `"type":"mysql"|"module":"mysql"`,
		config.DataTypeNginxModule:         `"type":"nginx"|"module":"nginx"`,
		config.DataTypeOsqueryModule:       `"type":"osquery"|"module":"osquery"`,
		config.DataTypePostgresqlModule:    `"type":"postgresql"|"module":"postgresql"`,
		config.DataTypeRedisModule:         `"type":"redis"|"module":"redis"`,
		config.DataTypeLinuxAgent:          `"type":"system"|"module":"system"`,
		config.DataTypeIisModule:           `"type":"iis"|"module":"iis"`,
		config.DataTypeTraefikModule:       `"type":"traefik"|"module":"traefik"`,
		config.DataTypeNatsModule:          `"type":"nats"|"module":"nats"`,
		config.DataTypeHaproxyModule:       `"type":"haproxy"|"module":"haproxy"`,
	}
)

type BeatsParser struct{}

func GetBeatsParser() *BeatsParser {
	beatsParserOnce.Do(func() {
		beatsParser = BeatsParser{}
	})
	return &beatsParser
}

func (p *BeatsParser) IdentifySource(log string) (config.DataType, error) {
	for logType, regp := range RegexspBeats {
		regExpCompiled, err := regexp.Compile(string(regp))
		if err != nil {
			return "", err
		}
		if regExpCompiled.MatchString(log) {
			return logType, nil
		}
	}
	return config.DataTypeGeneric, nil
}

func (p *BeatsParser) ProcessData(logBatch interface{}) (map[string][]string, error) {
	classifiedLogs := make(map[string][]string)
	batch, ok := logBatch.([]string)
	if !ok {
		return nil, fmt.Errorf("invalid log batch type")
	}
	for _, log := range batch {
		if logType, err := p.IdentifySource(log); err != nil {
			return nil, err
		} else {
			if logType != "" {
				validatedLog, _, err := validations.ValidateString(log, false)
				if err != nil {
					utils.Logger.ErrorF("error validating log: %s: %v", log, err)
					continue
				}
				classifiedLogs[string(logType)] = append(classifiedLogs[string(logType)], validatedLog)
			}
		}
	}
	return classifiedLogs, nil
}
