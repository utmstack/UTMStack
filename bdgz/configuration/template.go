package configuration

import (
	"github.com/utmstack/UTMStack/bdgz/constants"
	"github.com/utmstack/UTMStack/bdgz/schema"
	"github.com/utmstack/UTMStack/bdgz/utils"
	"github.com/utmstack/config-client-go/types"
)

var defaultSubscribeToEventTypes = []byte(`{
	"adcloud" : true ,
	"antiexploit" : true ,
	"aph" : true ,
	"av" : true ,
	"avc" : true ,
	"dp" : true ,
	"endpoint-moved-in" : true ,
	"endpoint-moved-out" : true ,
	"exchange-malware" : true ,
	"exchange-user-credentials" : true ,
	"fw" : true ,
	"hwid-change" : true ,
	"install" : false ,
	"modules" : false ,
	"network-monitor" : true ,
	"network-sandboxing" : true ,
	"new-incident" : true ,
	"ransomware-mitigation" : true ,
	"registration" : false ,
	"security-container-update-available" : true ,
	"supa-update-status" : true ,
	"sva" : true ,
	"sva-load" : true ,
	"task-status" : true ,
	"troubleshooting-activity" : true ,
	"uc" : true ,
	"uninstall" : false
}`)

// getTemplateTest returns a template that will be used to send the configuration to the PUSH API
func getTemplateSetPush(config types.ModuleGroup) schema.TemplateConfigSetPush {
	byteTemplate := schema.TemplateConfigSetPush{
		PARAMS: schema.Params{
			Status:      1,
			ServiceType: "cef",
			Servicesettings: schema.ServiceSettings{
				Url:                        "https://" + config.Configurations[2].ConfValue + ":" + constants.GetConnectorPort() + "/api",
				Authorization:              utils.GenerateAuthCode(config.Configurations[0].ConfValue),
				RequireValidSslCertificate: false,
			},
			SubscribeToEventTypes: defaultSubscribeToEventTypes,
		},
		JSONRPC: "2.0",
		Method:  "setPushEventSettings",
		ID:      "1",
	}
	return byteTemplate
}

// getTemplateTest returns a template that will be used to get the configuration of the PUSH API.
func getTemplateGet() schema.TemplateConfigGetPush {
	byteTemplate := schema.TemplateConfigGetPush{
		PARAMS:  []byte(`{}`),
		JSONRPC: "2.0",
		Method:  "getPushEventSettings",
		ID:      "3",
	}
	return byteTemplate
}

// getTemplateTest returns a template that will be used to send a test log
func getTemplateTest() schema.TemplateTestPush {
	byteTemplate := schema.TemplateTestPush{
		PARAMS: schema.ParamsTest{
			EventType: "av",
		},
		JSONRPC: "2.0",
		Method:  "sendTestPushEvent",
		ID:      "4",
	}
	return byteTemplate
}
