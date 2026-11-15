package upgrade

import (
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

type Config struct {
	AgentmanagerDSN string
	IntegrationsDir string
	TenantDomain    string
	RulesDir        string
	EngineDir       string
	FiltersDir      string
	VersionOldFile  string
}

// Run performs the one-time v11 -> v12 conversion. The installer's
// version-old.json decides v11 origin; jhi_user (dropped by CleanupOrphans)
// decides the conversion already ran, so reboots never re-upsert v11 values
// over edited v12 data.
func Run(db *gorm.DB, cnf Config) error {
	if !isV11FromFile(cnf.VersionOldFile) {
		return nil
	}
	if !tableExists(db, "jhi_user") {
		return nil
	}

	amdb, err := connectAgentmanager(cnf.AgentmanagerDSN)
	if err != nil {
		return fmt.Errorf("connect agentmanager database: %w", err)
	}
	defer amdb.Close()

	catcher.Info("v11 upgrade: converting existing data", nil)

	if err := ensureDefaultTenant(db, cnf.TenantDomain); err != nil {
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
	if err := convertPlugins(db, cnf.IntegrationsDir); err != nil {
		return err
	}
	if err := convertCorrelationRules(db, cnf.RulesDir, cnf.EngineDir); err != nil {
		return err
	}
	if err := convertLogstashFilters(db, cnf.FiltersDir, cnf.EngineDir); err != nil {
		return err
	}

	return nil
}
