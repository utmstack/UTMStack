package configuration

import (
	"log"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/utmstack/UTMStack/plugins/bitdefender/schema"
)

const DefaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"
const EndpointPush = "/v1.0/jsonrpc/push"

const UtmCertFileName = "utm.crt"
const UtmCertFileKey = "utm.key"

func GetConfig() *schema.PluginConfig {

	pCfg, e := helpers.PluginCfg[schema.PluginConfig]("com.utmstack")
	if e != nil {
		log.Fatalf("Failed to load plugin config: %v", e)
	}

	return pCfg

}