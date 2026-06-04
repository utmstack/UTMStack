package database

import (
	"errors"

	alerts_domain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	appconfig_domain "github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	asset_metrics_domain "github.com/utmstack/utmstack/backend/modules/asset_metrics/domain"
	audit_domain "github.com/utmstack/utmstack/backend/modules/audit/domain"
	collectors_domain "github.com/utmstack/utmstack/backend/modules/collectors/domain"
	correlation_domain "github.com/utmstack/utmstack/backend/modules/correlation/domain"
	datainput_domain "github.com/utmstack/utmstack/backend/modules/datainput/domain"
	iam_domain "github.com/utmstack/utmstack/backend/modules/iam/domain"
	ir_domain "github.com/utmstack/utmstack/backend/modules/incident_response/domain"
	incidents_domain "github.com/utmstack/utmstack/backend/modules/incidents/domain"
	indexpattern_domain "github.com/utmstack/utmstack/backend/modules/indexpattern/domain"
	logstash_domain "github.com/utmstack/utmstack/backend/modules/logstash/domain"
	notifications_domain "github.com/utmstack/utmstack/backend/modules/notifications/domain"
	arr_domain "github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gorm.io/gorm"

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
		audit_domain.AuditLog{},
		appconfig_domain.Config{},
		alerts_domain.UtmAlertTag{},
		alerts_domain.UtmAlertTagRule{},
		arr_domain.AlertResponseRule{},
		arr_domain.AlertResponseRuleExecution{},
		arr_domain.AlertResponseActionTemplate{},
		arr_domain.RuleTemplate{},
		correlation_domain.UtmRegexPattern{},
		correlation_domain.UtmTenantConfig{},
		correlation_domain.UtmDataTypes{},
		correlation_domain.UtmCorrelationRules{},
		datainput_domain.UtmDataInputStatus{},
		datainput_domain.UtmDataInputStatusCheckpoint{},
		logstash_domain.UtmLogstashFilterGroup{},
		logstash_domain.UtmLogstashFilter{},
		logstash_domain.UtmLogstashPipeline{},
		logstash_domain.UtmGroupLogstashPipelineFilters{},
		indexpattern_domain.UtmIndexPattern{},
		collectors_domain.UtmCollector{},
		asset_metrics_domain.UtmAssetMetric{},
		incidents_domain.UtmIncident{},
		incidents_domain.UtmIncidentAlert{},
		incidents_domain.UtmIncidentNote{},
		incidents_domain.UtmIncidentHistory{},
		ir_domain.UtmIncidentAction{},
		ir_domain.UtmIncidentActionCommand{},
		ir_domain.UtmIncidentJob{},
		ir_domain.UtmIncidentVariable{},
		notifications_domain.UtmNotification{},
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
		logger.Warn("could not create uuid-ossp extension: " + err.Error())
	}

	models := Models()
	if len(models) > 0 {
		logger.Info("running GORM AutoMigrate...")
		if err := db.AutoMigrate(models...); err != nil {
			return err
		}
	}

	if migrationsURL == "" {
		return nil
	}
	return runSQLMigrations(db, migrationsURL)
}

func runSQLMigrations(db *gorm.DB, migrationsURL string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "postgres", driver)
	if err != nil {
		return err
	}
	logger.Info("applying SQL migrations from " + migrationsURL + "...")
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	logger.Info("SQL migrations applied")
	return nil
}
