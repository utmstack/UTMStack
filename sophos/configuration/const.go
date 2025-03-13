package configuration

import "github.com/utmstack/UTMStack/sophos/utils"

const (
	CORRELATIONURL = "http://correlation:8080/v1/newlog"
	AUTHURL        = "https://id.sophos.com/api/v2/oauth2/token"
	WHOAMIURL      = "https://api.central.sophos.com/whoami/v1"
	CHECKCON       = "https://id.sophos.com"
)

func GetInternalKey() string {
	return utils.Getenv("INTERNAL_KEY")
}

func GetPanelServiceName() string {
	return utils.Getenv("PANEL_SERV_NAME")
}
