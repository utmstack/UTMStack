package main

import (
	"fmt"

	"github.com/utmstack/utmstack/backend/database"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDatabase(cfg *config) *gorm.DB {
	logger.Init(&logger.Config{
		MinLevel:    env.String("LOG_LEVEL", "INFO", false),
		ServiceName: "backend",
	})

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.dbHost, cfg.dbPort, cfg.dbUser, cfg.dbPass, cfg.dbName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Critical("failed to connect to database: " + err.Error())
		panic(err)
	}

	if err := database.MigrateDatabase(db, "file://migrations"); err != nil {
		logger.Critical("failed to migrate database: " + err.Error())
		panic(err)
	}

	return db
}
