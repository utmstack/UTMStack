package parser

import (
	"fmt"
	"regexp"
	"sync"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/agent/service/config"
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
	for DataType, regp := range RegexspBeats {
		regExpCompiled, err := regexp.Compile(string(regp))
		if err != nil {
			return "", err
		}
		if regExpCompiled.MatchString(log) {
			return DataType, nil
		}
	}
	return config.DataTypeGeneric, nil
}

func (p *BeatsParser) ProcessData(logMessage interface{}, datasource string, queue chan *go_sdk.Log) error {
	log, ok := logMessage.(string)
	if !ok {
		return fmt.Errorf("log is not of type string")
	}
	DataType, err := p.IdentifySource(log)
	if err != nil {
		return err
	}
	queue <- &go_sdk.Log{
		DataType:   string(DataType),
		DataSource: datasource,
		Raw:        log,
	}
	return nil
}
