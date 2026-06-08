package database

import (
	"errors"

	alerts_domain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	appconfig_domain "github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	compliance_domain "github.com/utmstack/utmstack/backend/modules/compliance/domain"
	dashboards_domain "github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	datasources_domain "github.com/utmstack/utmstack/backend/modules/datasources/domain"
	iam_domain "github.com/utmstack/utmstack/backend/modules/iam/domain"
	incidents_domain "github.com/utmstack/utmstack/backend/modules/incidents/domain"
	integrations_domain "github.com/utmstack/utmstack/backend/modules/integrations/domain"
	loganalyzer_domain "github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	notifications_domain "github.com/utmstack/utmstack/backend/modules/notifications/domain"
	opensearch_domain "github.com/utmstack/utmstack/backend/modules/opensearch/domain"
	arr_domain "github.com/utmstack/utmstack/backend/modules/soar/domain"
	"gorm.io/gorm"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Models lists every GORM-managed entity so AutoMigrate keeps the schema in
// sync with the Go struct tags. Append new domain types as you migrate each
// module.
func Models() []any {
	return []any{
		iam_domain.User{},
		iam_domain.Authority{},
		iam_domain.Permission{},
		iam_domain.AuthorityPermission{},
		iam_domain.UserAuthority{},
		iam_domain.APIKey{},
		iam_domain.RefreshToken{},
		iam_domain.IdentityProviderConfig{},
		audit_domain.AuditLog{},
		appconfig_domain.Config{},
		alerts_domain.UtmAlertTag{},
		alerts_domain.UtmAlertTagRule{},
		arr_domain.AlertResponseRule{},
		arr_domain.AlertResponseRuleExecution{},
		arr_domain.AlertResponseActionTemplate{},
		arr_domain.RuleTemplate{},
		arr_domain.UtmIncidentVariable{},
		arr_domain.UtmIncidentAction{},
		arr_domain.UtmIncidentActionCommand{},
		arr_domain.UtmIncidentJob{},
		compliance_domain.UtmComplianceReportConfig{},
		compliance_domain.UtmComplianceReportSchedule{},
		opensearch_domain.UtmIndexPattern{},
		integrations_domain.UtmModule{},
		incidents_domain.UtmIncident{},
		incidents_domain.UtmIncidentAlert{},
		incidents_domain.UtmIncidentNote{},
		incidents_domain.UtmIncidentHistory{},
		notifications_domain.UtmNotification{},
		datasources_domain.UtmAssetGroup{},
		datasources_domain.Datasource{},
		dashboards_domain.UtmDashboard{},
		dashboards_domain.UtmVisualization{},
		dashboards_domain.UtmDashboardVisualization{},
		loganalyzer_domain.UtmLogAnalyzerQuery{},
	}
}

// MigrateDatabase runs in two stages:
//
//  1. GORM AutoMigrate — creates/alters tables to match the Models() list.
//  2. golang-migrate over `migrationsURL` — applies SQL files for seed data,
//     composite indexes, and anything GORM struct tags can't express. Tracked
//     in schema_migrations and applied once.
//
// `migrationsURL` is a URL the file source understands; in dev and inside the
// container we pass "file://migrations" (cwd-relative) and the Dockerfile
// COPYs the SQL files next to the binary. Pass "" to skip SQL migrations.
func MigrateDatabase(db *gorm.DB, migrationsURL string) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		catcher.Warn("could not create uuid-ossp extension", map[string]any{"error": err.Error()})
	}

	if migrationsURL != "" {
		if err := runSQLMigrations(db, migrationsURL+"/pre", "schema_migrations_pre"); err != nil {
			return err
		}
	}

	models := Models()
	if len(models) > 0 {
		catcher.Info("running GORM AutoMigrate...", nil)
		if err := db.AutoMigrate(models...); err != nil {
			return err
		}
	}

	if migrationsURL == "" {
		return nil
	}
	return runSQLMigrations(db, migrationsURL, "schema_migrations")
}

func runSQLMigrations(db *gorm.DB, migrationsURL, table string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{MigrationsTable: table})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "postgres", driver)
	if err != nil {
		return err
	}
	catcher.Info("applying SQL migrations from "+migrationsURL+"...", nil)
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	catcher.Info("SQL migrations applied", nil)
	return nil
}
