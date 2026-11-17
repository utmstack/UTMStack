package database

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
)

// migratedFromV11 reports whether the database was migrated from v11 by reading the version-old.json file
func migratedFromV11() bool {
	data, err := os.ReadFile("/updates/version-old.json")
	if err != nil {
		return false
	}
	var v struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(v.Version), "v11.")
}

// v11TenantColumn is a list of SQL statements to add the tenant_id column to the agents and collectors tables if they are missing
var v11TenantColumn = []string{
	`ALTER TABLE IF EXISTS agents ADD COLUMN IF NOT EXISTS tenant_id text NOT NULL DEFAULT '` + config.DefaultTenantID + `'`,
	`ALTER TABLE IF EXISTS collectors ADD COLUMN IF NOT EXISTS tenant_id text NOT NULL DEFAULT '` + config.DefaultTenantID + `'`,
}

func MigrateDatabase() error {
	db := GetDB()
	if migratedFromV11() {
		for _, stmt := range v11TenantColumn {
			if err := db.GormDB().Exec(stmt).Error; err != nil {
				return err
			}
		}
	}
	return db.Migrate(
		&models.Agent{}, &models.AgentCommand{}, &models.LastSeen{}, &models.Collector{},
		&models.ConnectionKey{}, &models.Recovery{}, &models.RecoveryTarget{},
		&models.CollectorIntegrationConfig{}, &models.AgentConfigRevision{}, &models.AgentConfig{},
	)
}
