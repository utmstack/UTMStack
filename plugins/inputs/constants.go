package main

import (
	"os"

	go_sdk "github.com/threatwinds/go-sdk"
)

const panelConnectionKeyEndpoint = "%s/api/federation-service/token"
const panelAPIKeyHeader = "Utm-Internal-Key"
const proxyAPIKeyHeader = "Utm-Connection-Key"
const utmCertFileName = "utm.crt"
const utmCertFileKey = "utm.key"

func LogstashHost() string {
	cnf, e := go_sdk.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}
	return cnf.Logstash
}
