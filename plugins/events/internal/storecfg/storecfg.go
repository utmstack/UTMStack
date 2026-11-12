package storecfg

import (
	"fmt"

	"github.com/threatwinds/go-sdk/store/factory"
	"github.com/tidwall/gjson"
)

type Table struct {
	ConfigKey string
	Default   string
}

func ClickHousePrimary(cfg factory.Config) (factory.Config, error) {
	if cfg.Primary.Engine != factory.EngineClickHouse || cfg.Primary.ClickHouse == nil {
		return factory.Config{}, fmt.Errorf("a ClickHouse primary store is required")
	}
	cfg.Secondary = nil
	return cfg, nil
}

func New(storeSection, clickhouseSection gjson.Result, tables map[string]Table) (factory.Config, error) {
	if storeSection.Exists() {
		return factory.Parse([]byte(storeSection.Raw))
	}

	configuredTables := make(map[string]string, len(tables))
	for dataset, table := range tables {
		name := clickhouseSection.Get(table.ConfigKey).String()
		if name == "" {
			name = table.Default
		}
		configuredTables[dataset] = name
	}
	return factory.Config{Primary: factory.Engine{
		Engine: factory.EngineClickHouse,
		ClickHouse: &factory.ClickHouse{
			Addr:           []string{clickhouseSection.Get("host").String() + ":" + clickhouseSection.Get("port").String()},
			Database:       clickhouseSection.Get("database").String(),
			Username:       clickhouseSection.Get("user").String(),
			Password:       clickhouseSection.Get("password").String(),
			Tables:         configuredTables,
			TenantColumn:   "tenantId",
			TimeColumn:     "@timestamp",
			DataTypeColumn: "dataType",
		},
	}}, nil
}
