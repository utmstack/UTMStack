package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/threatwinds/go-sdk/catcher"
	// _ "github.com/utmstack/utmstack/backend/docs"
	"github.com/utmstack/utmstack/backend/pkg/env"
)

// @Title UTMStack Backend Service
// @Version 1.0
// @Description UTMStack backend API — manages IAM, modules, flows, incidents, compliance and audit data.
// @BasePath /api/v1
// @Schemes https http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	env.LoadDotEnv()
	cfg := loadConfig()

	db := initDatabase(cfg)

	catcher.Info(fmt.Sprintf(
		"config loaded: appPort=%d devMode=%t serverName=%q jwtIssuer=%q tfaEnabled=%t (NOTE: backend does NOT read .env files; values come from process environment only)",
		cfg.appPort, cfg.devMode, cfg.serverName, cfg.jwtIssuer, cfg.tfaEnabled,
	), nil)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	appCtx, appCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer appCancel()

	modules := initModules(db, cfg)

	adminEmail := env.String("UTMSTACK_ADMIN_EMAIL", "admin", false)
	created, err := modules.tenant.GetBootstrapUsecase().EnsureDefaultTenant(
		appCtx, adminEmail,
		env.String("UTMSTACK_ADMIN_PASSWORD", "", false),
		env.String("UTMSTACK_DEFAULT_DOMAIN", "", false))
	if err != nil {
		_ = catcher.Error("failed to create the default tenant", err, nil)
		panic(err)
	}
	if created {
		// Never the password: the installer already showed it once, and these
		// logs are kept.
		catcher.Info(fmt.Sprintf("created the default tenant and its administrator %q", adminEmail), nil)
	}

	// TODO(scaling): these Start calls launch periodic jobs, and every replica
	// runs all of them. That is fine today because the backend runs as one
	// replica, and it is not the layer where the problem should be solved
	// anyway: the direction is for plugins to own the scheduling and drive the
	// backend, which leaves it request-driven with nothing periodic to
	// coordinate.
	//
	// Whatever survives that move needs the fix SOAR and the compliance report
	// scheduler already have — an atomic claim on the row being worked
	// (ClaimPending / ClaimDue). That is correct under polling, under an
	// external trigger and with N replicas, because it does not depend on how
	// the work is kicked off. Coordinating at the job level instead only
	// serialises replicas that could have processed different rows.
	//
	// Cache reloads (filter_store, rule_store, flow_store, reloadCoverage) are
	// the exception: they rebuild in-process state and every replica has to run
	// them.
	modules.audit.Start(appCtx)
	modules.notifications.Start(appCtx)
	modules.iam.Start(appCtx)

	modules.billing.Start(appCtx)
	modules.eventProcessing.Start(appCtx)
	modules.compliance.Start(appCtx)
	modules.integrations.Start(appCtx)
	modules.datasources.Start(appCtx)
	if err := modules.soar.Start(appCtx); err != nil {
		_ = catcher.Error("soar flow bootstrap failed", err, nil)
	}

	engine := initHTTPServer(cfg)
	registerRoutes(engine, modules, cfg)

	// startServer blocks until appCtx is cancelled (signal received).
	startServer(engine, cfg, appCtx)
	modules.audit.Stop()
}
