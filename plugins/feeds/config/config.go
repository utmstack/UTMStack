package config

import (
	"github.com/threatwinds/go-sdk/plugins"
)

type TWConfig struct {
	InternalKey        string
	BackendURL         string
	ThreadWindsURL     string
	OpenSearchHost     string
	OpenSearchPort     string
	OpenSearchUser     string
	OpenSearchPassword string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
}

func GetTWConfig() (*TWConfig, error) {
	utmCfg := plugins.PluginCfg("com.utmstack")
	osCfg := plugins.PluginCfg("org.opensearch").Get("opensearch")
	pgCfg := utmCfg.Get("postgresql")

	cfg := &TWConfig{
		InternalKey:        utmCfg.Get("internalKey").String(),
		BackendURL:         utmCfg.Get("backend").String(),
		ThreadWindsURL:     GetThreadWindsURL(),
		OpenSearchHost:     osCfg.Get("host").String(),
		OpenSearchPort:     osCfg.Get("port").String(),
		OpenSearchUser:     osCfg.Get("user").String(),
		OpenSearchPassword: osCfg.Get("password").String(),
		DBHost:             pgCfg.Get("server").String(),
		DBPort:             pgCfg.Get("port").String(),
		DBUser:             pgCfg.Get("user").String(),
		DBPassword:         pgCfg.Get("password").String(),
		DBName:             pgCfg.Get("database").String(),
	}

	return cfg, nil
}
