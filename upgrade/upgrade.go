package upgrade

import (
	"database/sql"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Run performs the one-time v11 -> v12 conversion. Detection is v11-table
// presence; idempotency comes from each converter's upserts, and the orphan
// cleanup drops the v11 tables so a re-detection can never happen again.
func Run(db *gorm.DB, agentmanagerDSN, yamlDir, tenantDomain, rulesDir, pipelineDir, pipelinesDir string) error {
	if !isV11(db) {
		return nil
	}

	amdb, err := connectAgentmanager(agentmanagerDSN)
	if err != nil {
		return fmt.Errorf("connect agentmanager database: %w", err)
	}
	defer amdb.Close()

	catcher.Info("v11 upgrade: converting existing data", nil)

	if err := ensureDefaultTenant(db, tenantDomain); err != nil {
		return err
	}
	if err := convertUsers(db); err != nil {
		return err
	}
	if err := convertAppConfig(db); err != nil {
		return err
	}
	tagIDs, err := convertAlertTags(db)
	if err != nil {
		return err
	}
	if _, err := convertAlertTagRules(db, tagIDs); err != nil {
		return err
	}
	if err := convertIncidents(db); err != nil {
		return err
	}
	if err := convertFederation(db, amdb); err != nil {
		return err
	}
	if err := convertAPIKeys(db); err != nil {
		return err
	}
	if err := convertPlugins(db, yamlDir); err != nil {
		return err
	}
	if err := convertCorrelationRules(db, rulesDir, pipelineDir); err != nil {
		return err
	}
	if err := convertLogstashFilters(db, pipelinesDir, pipelineDir); err != nil {
		return err
	}

	return nil
}

func isV11(db *gorm.DB) bool {
	var exists bool
	db.Raw(`SELECT to_regclass('jhi_user') IS NOT NULL`).Scan(&exists)
	return exists
}

func connectAgentmanager(dsn string) (*sql.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return sqlDB, nil
}
