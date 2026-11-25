package config

import (
	"github.com/utmstack/UTMStack/integrations/bitdefender/schema"
	"github.com/utmstack/UTMStack/integrations/bitdefender/utils"
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

func getTemplateSetPush(config BDGZModuleConfig) schema.TemplateConfigSetPush {
	byteTemplate := schema.TemplateConfigSetPush{
		Params: schema.Params{
			Status:      1,
			ServiceType: "cef",
			ServiceSettings: schema.ServiceSettings{
				Url:                        "https://" + config.MasterIp + ":" + BitdefenderGZPort + "/api",
				Authorization:              utils.GenerateAuthCode(config.ConnectionKey),
				RequireValidSslCertificate: false,
			},
			SubscribeToEventTypes: defaultSubscribeToEventTypes,
		},
		JsonRpc: "2.0",
		Method:  "setPushEventSettings",
		Id:      "1",
	}
	return byteTemplate
}

func getTemplateGet() schema.TemplateConfigGetPush {
	byteTemplate := schema.TemplateConfigGetPush{
		Params:  []byte(`{}`),
		JsonRpc: "2.0",
		Method:  "getPushEventSettings",
		Id:      "3",
	}
	return byteTemplate
}

func getTemplateTest() schema.TemplateTestPush {
	byteTemplate := schema.TemplateTestPush{
		Params: schema.ParamsTest{
			EventType: "av",
		},
		JsonRpc: "2.0",
		Method:  "sendTestPushEvent",
		Id:      "4",
	}
	return byteTemplate
}
