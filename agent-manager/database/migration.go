package database

import (
	"github.com/utmstack/UTMStack/agent-manager/models"
)

// v12 adds NOT-NULL columns without a default to tables that carry rows in a
// v11 agentmanager DB; GORM cannot add those to a non-empty table, so main.go
// would exit(1) and agent-manager crash-loops. Pre-create them with a safe
// non-volatile default before Migrate. v11 agents keep the empty tenant (their
// re-enrolment is a separate follow-up); the column only needs to exist so the
// schema matches the v12 model.
var v11DefensiveColumns = []string{
	`ALTER TABLE IF EXISTS agents ADD COLUMN IF NOT EXISTS tenant_id text NOT NULL DEFAULT ''`,
	`ALTER TABLE IF EXISTS collectors ADD COLUMN IF NOT EXISTS tenant_id text NOT NULL DEFAULT ''`,
}

func MigrateDatabase() error {
	db := GetDB()
	for _, stmt := range v11DefensiveColumns {
		if err := db.GormDB().Exec(stmt).Error; err != nil {
			return err
		}
	}
	return db.Migrate(
		&models.Agent{}, &models.AgentCommand{}, &models.LastSeen{}, &models.Collector{},
		&models.ConnectionKey{}, &models.Recovery{}, &models.RecoveryTarget{},
		&models.CollectorIntegrationConfig{}, &models.AgentConfigRevision{}, &models.AgentConfig{},
	)
}
