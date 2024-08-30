package config

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/utils"
)

const REPLACE_KEY string = ""

const (
	AgentVersion                 = "10.6.0"
	DEPEND_URL                   = "https://%s/dependencies/agent?version=%s&os=%s&type=%s"
	AGENTMANAGERPORT      string = "9000"
	CORRELATIONLOGSPORT   string = "50051"
	MaxConnectionTime            = 120 * time.Second
	InitialReconnectDelay        = 10 * time.Second
	MaxReconnectDelay            = 120 * time.Second

	SERV_NAME            = "UTMStackAgent"
	SERV_LOG             = "utmstack_agent.log"
	ModulesServName      = "UTMStackModulesLogsCollector"
	WinServName          = "UTMStackWindowsLogsCollector"
	ModulesLockName      = "utmstack_modules_collector.lock"
	WinLockName          = "utmstack_windows_collector.lock"
	CollectorFileNameOld = "log-collector-configuration.json"
	CollectorFileName    = "log-collector-config.json"
	UUIDFileName         = "uuid.yml"
	DEPEND_ZIP_LABEL     = "depend_zip"
	DEPEND_SERVICE_LABEL = "service"
	BatchToSend          = 5
	PortRangeMin         = "7000"
	PortRangeMax         = "9000"
	RetentionConfigFile  = "retention.json"
	LogsDBFile           = "logs.db"
)

type DataType string

var (
	DataTypeWindowsAgent        DataType = "wineventlog"
	DataTypeSyslog              DataType = "syslog"
	DataTypeVmware              DataType = "vmware-esxi"
	DataTypeLinuxAgent          DataType = "linux"
	DataTypeEset                DataType = "antivirus-esmc-eset"
	DataTypeKaspersky           DataType = "antivirus-kaspersky"
	DataTypeTraefikModule       DataType = "traefik"
	DataTypeMongodbModule       DataType = "mongodb"
	DataTypeMysqlModule         DataType = "mysql"
	DataTypePostgresqlModule    DataType = "postgresql"
	DataTypeRedisModule         DataType = "redis"
	DataTypeElasticsearchModule DataType = "elasticsearch"
	DataTypeKafkaModule         DataType = "kafka"
	DataTypeKibanaModule        DataType = "kibana"
	DataTypeLogstashModule      DataType = "logstash"
	DataTypeCiscoAsa            DataType = "firewall-cisco-asa"
	DataTypeCiscoMeraki         DataType = "firewall-meraki"
	DataTypeFortinet            DataType = "firewall-fortigate-traffic"
	DataTypePaloalto            DataType = "firewall-paloalto"
	DataTypeMikrotik            DataType = "firewall-mikrotik"
	DataTypeCiscoFirepower      DataType = "firewall-cisco-firepower"
	DataTypeSophosXG            DataType = "firewall-sophos-xg"
	DataTypeCiscoSwitch         DataType = "cisco-switch"
	DataTypeSonicwall           DataType = "firewall-sonicwall"
	DataTypeNatsModule          DataType = "nats"
	DataTypeDeceptivebytes      DataType = "deceptive-bytes"
	DataTypeOsqueryModule       DataType = "osquery"
	DataTypeLinuxAuditdModule   DataType = "auditd"
	DataTypeHaproxyModule       DataType = "haproxy"
	DataTypeNginxModule         DataType = "nginx"
	DataTypeIisModule           DataType = "iis"
	DataTypeApacheModule        DataType = "apache"
	DataTypeSentinelOne         DataType = "antivirus-sentinel-one"
	DataTypeCiscoGeneric        DataType = "cisco"
	DataTypeMacOs               DataType = "macos"
	DataTypeGeneric             DataType = "generic"
	DataTypeNetflow             DataType = "netflow"
	DataTypeAix                 DataType = "ibm-aix"
	DataTypePfsense             DataType = "firewall-pfsense"
	DataTypeFortiweb            DataType = "firewall-fortiweb"
)

type ProtoPort struct {
	UDP string
	TCP string
}

var (
	ProtoPorts = map[DataType]ProtoPort{
		DataTypeSyslog:         {UDP: "7014", TCP: "7014"},
		DataTypeVmware:         {UDP: "7002", TCP: "7002"},
		DataTypeEset:           {UDP: "7003", TCP: "7003"},
		DataTypeKaspersky:      {UDP: "7004", TCP: "7004"},
		DataTypeCiscoGeneric:   {UDP: "514", TCP: "1470"},
		DataTypeFortinet:       {UDP: "7005", TCP: "7005"},
		DataTypePaloalto:       {UDP: "7006", TCP: "7006"},
		DataTypeMikrotik:       {UDP: "7007", TCP: "7007"},
		DataTypeSophosXG:       {UDP: "7008", TCP: "7008"},
		DataTypeSonicwall:      {UDP: "7009", TCP: "7009"},
		DataTypeDeceptivebytes: {UDP: "7010", TCP: "7010"},
		DataTypeSentinelOne:    {UDP: "7012", TCP: "7012"},
		DataTypeMacOs:          {UDP: "7015", TCP: "7015"},
		DataTypeAix:            {UDP: "7016", TCP: "7016"},
		DataTypePfsense:        {UDP: "7017", TCP: "7017"},
		DataTypeFortiweb:       {UDP: "7018", TCP: "7018"},
		DataTypeNetflow:        {UDP: "2055", TCP: ""},
	}

	ProhibitedPortsChange = []DataType{DataTypeCiscoGeneric, DataTypeNetflow}
)

func GetCertPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.crt")
}

func GetKeyPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.key")
}

func GetCaPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "certs", "ca.crt")
}

func GetCollectorConfigPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, CollectorFileName)
}

func GetCollectorConfigPathOld() string {
	path := utils.GetMyPath()
	return filepath.Join(path, CollectorFileNameOld)
}

func GetAgentBin() string {
	var bin string
	switch runtime.GOOS {
	case "windows":
		bin = "utmstack_agent_service.exe"
	case "linux":
		bin = "utmstack_agent_service"
	}
	return bin
}

func ValidateModuleType(typ string) string {
	switch DataType(typ) {
	case DataTypeSyslog, DataTypeVmware, DataTypeEset, DataTypeKaspersky, DataTypeFortinet, DataTypePaloalto,
		DataTypeMikrotik, DataTypeSophosXG, DataTypeSonicwall, DataTypeSentinelOne, DataTypeCiscoGeneric, DataTypeMacOs,
		DataTypeDeceptivebytes, DataTypeAix, DataTypePfsense, DataTypeFortiweb:
		return "syslog"
	case DataTypeNetflow:
		return "netflow"
	case DataTypeWindowsAgent, DataTypeLinuxAgent, DataTypeTraefikModule, DataTypeMongodbModule, DataTypeMysqlModule, DataTypePostgresqlModule,
		DataTypeRedisModule, DataTypeElasticsearchModule, DataTypeKafkaModule, DataTypeKibanaModule, DataTypeLogstashModule, DataTypeNatsModule,
		DataTypeOsqueryModule, DataTypeLinuxAuditdModule, DataTypeHaproxyModule, DataTypeNginxModule, DataTypeIisModule, DataTypeApacheModule:
		return "beats"
	default:
		return "nil"
	}
}

func GetDependenPaths() []string {
	path := utils.GetMyPath()
	self := "utmstack_updater_self"
	if runtime.GOOS == "windows" {
		self += ".exe"
	}
	return []string{
		filepath.Join(path, "beats"),
		filepath.Join(path, "templates"),
		filepath.Join(path, self),
	}
}

func GetDownloadFilePath(typ string, subfix string) string {
	path := utils.GetMyPath()
	switch typ {
	case DEPEND_SERVICE_LABEL:
		switch runtime.GOOS {
		case "windows":
			return filepath.Join(path, fmt.Sprintf("utmstack_agent_service%s.exe", subfix))
		case "linux":
			return filepath.Join(path, fmt.Sprintf("utmstack_agent_service%s", subfix))
		}
	case DEPEND_ZIP_LABEL:
		return filepath.Join(path, "dependencies.zip")
	}
	return ""
}

func GetVersionPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "version.json")
}

func GetVersionOldPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "versions.json")
}

func GetSelfUpdaterPath() string {
	path := utils.GetMyPath()
	if runtime.GOOS == "windows" {
		return filepath.Join(path, "utmstack_updater_self.exe")
	}
	return filepath.Join(path, "utmstack_updater_self")
}

func GetDependenciesServices() []string {
	services := []string{"UTMStackModulesLogsCollector"}
	if runtime.GOOS == "windows" {
		services = append(services, "UTMStackWindowsLogsCollector")
	}
	return services
}
