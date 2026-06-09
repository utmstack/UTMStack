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

	if err := modules.opensearchGateway.Start(appCtx); err != nil {
		_ = catcher.Error("opensearch module failed to start", err, nil)
		panic(err)
	}

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
}
