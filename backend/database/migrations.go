package database

import (
	"errors"
	"fmt"

	adaudit_domain "github.com/utmstack/utmstack/backend/modules/adaudit/domain"
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
	arr_domain "github.com/utmstack/utmstack/backend/modules/soar/domain"
	socai_domain "github.com/utmstack/utmstack/backend/modules/socai/domain"
	tenant_domain "github.com/utmstack/utmstack/backend/modules/tenant/domain"
	"github.com/utmstack/utmstack/backend/pkg/joblease"
	"gorm.io/gorm"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Models() []any {
	return []any{
		iam_domain.User{},
		tenant_domain.Tenant{},
		iam_domain.Role{},
		iam_domain.Permission{},
		iam_domain.RolePermission{},
		iam_domain.UserRole{},
		iam_domain.APIKey{},
		iam_domain.RefreshToken{},
		iam_domain.IdentityProviderConfig{},
		iam_domain.IdentityProviderGroupMapping{},
		iam_domain.UserChallenge{},
		iam_domain.TfaFactor{},
		audit_domain.AuditLog{},
		socai_domain.AIUsage{},
		joblease.Lease{},
		appconfig_domain.Config{},
		alerts_domain.AlertTag{},
		alerts_domain.AlertTagRule{},
		arr_domain.SoarExecution{},
		arr_domain.SoarFlowRun{},
		arr_domain.SoarExecutionEdge{},
		arr_domain.SoarVariable{},
		compliance_domain.ReportSchedule{},
		compliance_domain.Report{},
		compliance_domain.ReportScore{},
		integrations_domain.Integration{},
		incidents_domain.Incident{},
		incidents_domain.IncidentAlert{},
		incidents_domain.IncidentNote{},
		incidents_domain.IncidentHistory{},
		notifications_domain.UtmNotification{},
		datasources_domain.Datasource{},
		dashboards_domain.Dashboard{},
		dashboards_domain.Visualization{},
		loganalyzer_domain.SavedQuery{},
		adaudit_domain.ADUser{},
	}
}

func MigrateDatabase(db *gorm.DB, migrationsURL string) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		catcher.Warn("could not create uuid-ossp extension", map[string]any{"error": err.Error()})
	}

	// ponytail: v11 soar_executions rows predate the DAG columns. GORM
	// AutoMigrate can't add NOT NULL columns without a default when the table
	// already has rows, so we pre-create them here with DEFAULT ''. New inserts
	// override the default; legacy rows keep empty strings, which is what the
	// DAG dispatcher already tolerates for terminal history.
	for _, stmt := range []string{
		`ALTER TABLE IF EXISTS soar_executions ADD COLUMN IF NOT EXISTS node_id  varchar(150) NOT NULL DEFAULT ''`,
		`ALTER TABLE IF EXISTS soar_executions ADD COLUMN IF NOT EXISTS kind     varchar(20)  NOT NULL DEFAULT ''`,
		`ALTER TABLE IF EXISTS soar_executions ADD COLUMN IF NOT EXISTS executor varchar(60)  NOT NULL DEFAULT ''`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	models := Models()
	if len(models) > 0 {
		catcher.Info("running GORM AutoMigrate...", nil)
		if err := db.AutoMigrate(models...); err != nil {
			return err
		}
		catcher.Info("GORM AutoMigrate complete", nil)
	}

	if migrationsURL == "" {
		return nil
	}
	return runSQLMigrations(db, migrationsURL, "schema_migrations", "post-AutoMigrate")
}

func runSQLMigrations(db *gorm.DB, migrationsURL, table, stage string) error {
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
	if err := applyUp(m, stage); err != nil {
		var dirty migrate.ErrDirty
		if !errors.As(err, &dirty) {
			return err
		}
		// A previous run started version dirty.Version but failed mid-way.
		// All our migrations are idempotent (IF EXISTS / ON CONFLICT), so it is
		// safe to force the version back one step and retry.
		catcher.Warn(fmt.Sprintf("%s SQL migration version %d is dirty — forcing back and retrying", stage, dirty.Version), nil)
		if ferr := m.Force(dirty.Version - 1); ferr != nil {
			return fmt.Errorf("could not force migration version after dirty state: %w", ferr)
		}
		if rerr := applyUp(m, stage); rerr != nil {
			return rerr
		}
	}
	return nil
}

func applyUp(m *migrate.Migrate, stage string) error {
	switch err := m.Up(); {
	case err == nil:
		catcher.Info(stage+" SQL migrations applied", nil)
		return nil
	case errors.Is(err, migrate.ErrNoChange):
		catcher.Info(stage+" SQL migrations already up to date", nil)
		return nil
	default:
		return err
	}
}
