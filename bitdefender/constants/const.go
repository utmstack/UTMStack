package constants

import "github.com/utmstack/UTMStack/bitdefender/utils"

const EndpointPush = "/v1.0/jsonrpc/push"

const URL_CHECK_CONNECTION = "https://cloud.gravityzone.bitdefender.com"

func GetInternalKey() string {
	return utils.Getenv("INTERNAL_KEY")
}

func GetPanelServiceName() string {
	return utils.Getenv("PANEL_SERV_NAME")
}

func GetSyslogHost() string {
	return utils.Getenv("SYSLOG_HOST")
}

func GetSyslogProto() string {
	return utils.Getenv("SYSLOG_PROTOCOL")
}

func GetSyslogPort() string {
	return utils.Getenv("SYSLOG_PORT")
}

func GetConnectorPort() string {
	return utils.Getenv("CONNECTOR_PORT")
}
