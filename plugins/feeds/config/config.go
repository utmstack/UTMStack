package config

import (
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store/factory"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/storecfg"
)

const defaultAlertsTable = "alerts"

type TWConfig struct {
	InternalKey string
	BackendURL  string

	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseDatabase string
	ClickHouseUser     string
	ClickHousePassword string
	AlertsTable        string
	Store              factory.Config
}

func GetTWConfig() (*TWConfig, error) {
	utmCfg := plugins.PluginCfg("com.utmstack")
	chCfg := plugins.PluginCfg("clickhouse")

	table := chCfg.Get("alertsTable").String()
	if table == "" {
		table = defaultAlertsTable
	}
	storeConfig, err := storecfg.New(plugins.PluginCfg("store"), chCfg, map[string]storecfg.Table{
		"alerts": {ConfigKey: "alertsTable", Default: defaultAlertsTable},
	})
	if err != nil {
		return nil, err
	}

	cfg := &TWConfig{
		InternalKey: utmCfg.Get("internalKey").String(),
		BackendURL:  utmCfg.Get("backend").String(),

		ClickHouseHost:     chCfg.Get("host").String(),
		ClickHousePort:     chCfg.Get("port").String(),
		ClickHouseDatabase: chCfg.Get("database").String(),
		ClickHouseUser:     chCfg.Get("user").String(),
		ClickHousePassword: chCfg.Get("password").String(),
		AlertsTable:        table,
		Store:              storeConfig,
	}

	return cfg, nil
}
