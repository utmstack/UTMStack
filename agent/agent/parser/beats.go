package parser

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/agent/config"
)

var (
	beatsParser     = BeatsParser{}
	beatsParserOnce sync.Once
	RegexspBeats    = map[config.LogType]string{
		config.LogTypeApacheModule:        `"type":"apache"|"module":"apache"`,
		config.LogTypeLinuxAuditdModule:   `"type":"auditd"|"module":"auditd"`,
		config.LogTypeElasticsearchModule: `"type":"elasticsearch"|"module":"elasticsearch"`,
		config.LogTypeKafkaModule:         `"type":"kafka"|"module":"kafka"`,
		config.LogTypeKibanaModule:        `"type":"kibana"|"module":"kibana"`,
		config.LogTypeLogstashModule:      `"type":"logstash"|"module":"logstash"`,
		config.LogTypeMongodbModule:       `"type":"mongodb"|"module":"mongodb"`,
		config.LogTypeMysqlModule:         `"type":"mysql"|"module":"mysql"`,
		config.LogTypeNginxModule:         `"type":"nginx"|"module":"nginx"`,
		config.LogTypeOsqueryModule:       `"type":"osquery"|"module":"osquery"`,
		config.LogTypePostgresqlModule:    `"type":"postgresql"|"module":"postgresql"`,
		config.LogTypeRedisModule:         `"type":"redis"|"module":"redis"`,
		config.LogTypeLinuxAgent:          `"type":"system"|"module":"system"`,
		config.LogTypeIisModule:           `"type":"iis"|"module":"iis"`,
		config.LogTypeTraefikModule:       `"type":"traefik"|"module":"traefik"`,
		config.LogTypeNatsModule:          `"type":"nats"|"module":"nats"`,
		config.LogTypeHaproxyModule:       `"type":"haproxy"|"module":"haproxy"`,
	}
)

type BeatsParser struct{}

func GetBeatsParser() *BeatsParser {
	beatsParserOnce.Do(func() {
		beatsParser = BeatsParser{}
	})
	return &beatsParser
}

func (p *BeatsParser) IdentifySource(log string) (string, error) {
	for logType, regp := range RegexspBeats {
		regExpCompiled, err := regexp.Compile(string(regp))
		if err != nil {
			return "", err
		}
		if regExpCompiled.MatchString(log) {
			return logType.DataType, nil
		}
	}
	return config.LogTypeGeneric.DataType, nil
}

func (p *BeatsParser) ProcessData(logMessage interface{}, datasource string, queue chan *plugins.Log) error {
	log, ok := logMessage.(string)
	if !ok {
		return fmt.Errorf("log is not of type string")
	}
	logType, err := p.IdentifySource(log)
	if err != nil {
		return err
	}
	queue <- &plugins.Log{
		DataType:   string(logType),
		DataSource: datasource,
		Raw:        log,
	}
	return nil
}
