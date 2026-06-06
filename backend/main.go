package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	// _ "github.com/utmstack/utmstack/backend/docs"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"github.com/utmstack/utmstack/backend/pkg/logger"
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

	logger.Info(fmt.Sprintf(
		"config loaded: appPort=%d devMode=%t serverName=%q jwtIssuer=%q tfaEnabled=%t (NOTE: backend does NOT read .env files; values come from process environment only)",
		cfg.appPort, cfg.devMode, cfg.serverName, cfg.jwtIssuer, cfg.tfaEnabled,
	))
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	appCtx, appCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer appCancel()

	modules := initModules(db, cfg)

	if err := modules.opensearchGateway.Start(appCtx); err != nil {
		logger.Critical("opensearch module failed to start: " + err.Error())
		panic(err)
	}

	modules.alerts.Start(appCtx)
	modules.eventProcessing.Start(appCtx)
	modules.compliance.Start(appCtx)
	modules.integrations.Start(appCtx)
	if sch := modules.networkScan.Scheduler(); sch != nil {
		go sch.Start(appCtx)
	}

	engine := initHTTPServer(cfg)
	registerRoutes(engine, modules, cfg)

	// startServer blocks until appCtx is cancelled (signal received).
	startServer(engine, cfg, appCtx)
}
