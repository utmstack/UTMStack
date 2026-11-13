package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbConnectMaxAttempts = 30
	dbConnectMaxBackoff  = 30 * time.Second
)

func gormLogLevel(s string) logger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}

func initDatabase(cfg *config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.dbHost, cfg.dbPort, cfg.dbUser, cfg.dbPass, cfg.dbName,
	)
	agentmanagerDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.dbHost, cfg.dbPort, cfg.dbUser, cfg.dbPass, cfg.dbAgent,
	)

	db := connectWithRetry(dsn, gormLogLevel(cfg.dbLogLevel))

	up := database.UpgradeOptions{
		AgentmanagerDSN: agentmanagerDSN,
		IntegrationsDir: cfg.integrationsDir,
		TenantDomain:    cfg.tenantDomain,
		RulesDir:        cfg.rulesDir,
		PipelineDir:     cfg.pipelineDir,
		PipelinesDir:    cfg.pipelinesDir,
	}
	if err := database.MigrateDatabase(db, "file://migrations/pre", "file://migrations/post", up); err != nil {
		_ = catcher.Error("failed to migrate database", err, nil)
		panic(err)
	}
	catcher.Info("✅ database ready — migrations complete, starting backend", nil)

	return db
}

func connectWithRetry(dsn string, logLevel logger.LogLevel) *gorm.DB {
	backoff := 2 * time.Second
	for attempt := 1; ; attempt++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
		})
		if err == nil {
			sqlDB, derr := db.DB()
			if derr != nil {
				err = derr
			} else if perr := sqlDB.Ping(); perr != nil {
				err = perr
			} else {
				if attempt > 1 {
					catcher.Info(fmt.Sprintf("connected to database after %d attempts", attempt), nil)
				}
				return db
			}
		}

		if attempt >= dbConnectMaxAttempts {
			_ = catcher.Error(fmt.Sprintf("failed to connect to database after %d attempts", attempt), err, nil)
			panic(err)
		}

		catcher.Warn(fmt.Sprintf("database not ready (attempt %d/%d); retrying in %s",
			attempt, dbConnectMaxAttempts, backoff), map[string]any{"error": err.Error()})
		time.Sleep(backoff)
		if backoff < dbConnectMaxBackoff {
			backoff *= 2
			if backoff > dbConnectMaxBackoff {
				backoff = dbConnectMaxBackoff
			}
		}
	}
}
