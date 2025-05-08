package configuration

import "github.com/utmstack/UTMStack/sophos/utils"

const (
	AUTHURL          = "https://id.sophos.com/api/v2/oauth2/token"
	WHOAMIURL        = "https://api.central.sophos.com/whoami/v1"
	CHECKCON         = "https://id.sophos.com"
	LogstashEndpoint = "http://%s:%s"
	UTMLogSeparator  = "<utm-log-separator>"
)

func GetInternalKey() string {
	return utils.Getenv("INTERNAL_KEY")
}

func GetPanelServiceName() string {
	return utils.Getenv("PANEL_SERV_NAME")
}

func GetLogstashHost() string {
	return utils.Getenv("UTM_LOGSTASH_HOST")
}

func GetLogstashPort() string {
	return utils.Getenv("UTM_LOGSTASH_PORT")
}
