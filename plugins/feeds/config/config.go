package config

import (
	"github.com/threatwinds/go-sdk/plugins"
)

const defaultAlertsTable = "alerts"

type TWConfig struct {
	InternalKey    string
	BackendURL     string
	ThreadWindsURL string

	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseDatabase string
	ClickHouseUser     string
	ClickHousePassword string
	AlertsTable        string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func GetTWConfig() (*TWConfig, error) {
	utmCfg := plugins.PluginCfg("com.utmstack")
	chCfg := plugins.PluginCfg("clickhouse")
	pgCfg := utmCfg.Get("postgresql")

	table := chCfg.Get("alertsTable").String()
	if table == "" {
		table = defaultAlertsTable
	}

	cfg := &TWConfig{
		InternalKey:    utmCfg.Get("internalKey").String(),
		BackendURL:     utmCfg.Get("backend").String(),
		ThreadWindsURL: GetThreadWindsURL(),

		ClickHouseHost:     chCfg.Get("host").String(),
		ClickHousePort:     chCfg.Get("port").String(),
		ClickHouseDatabase: chCfg.Get("database").String(),
		ClickHouseUser:     chCfg.Get("user").String(),
		ClickHousePassword: chCfg.Get("password").String(),
		AlertsTable:        table,

		DBHost:     pgCfg.Get("server").String(),
		DBPort:     pgCfg.Get("port").String(),
		DBUser:     pgCfg.Get("user").String(),
		DBPassword: pgCfg.Get("password").String(),
		DBName:     pgCfg.Get("database").String(),
	}

	return cfg, nil
}
