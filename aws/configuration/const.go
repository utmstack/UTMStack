package configuration

import "github.com/utmstack/UTMStack/aws/utils"

const (
	CORRELATIONURL = "http://correlation:8080/v1/newlog"
)

func GetInternalKey() string {
	return utils.Getenv("INTERNAL_KEY")
}

func GetPanelServiceName() string {
	return utils.Getenv("PANEL_SERV_NAME")
}
