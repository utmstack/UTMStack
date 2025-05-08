package configuration

import "github.com/utmstack/UTMStack/aws/utils"

const (
	URL_CHECK_CONNECTION = "https://sts.amazonaws.com"
	LogstashEndpoint     = "http://%s:%s"
	UTMLogSeparator      = "<utm-log-separator>"
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
