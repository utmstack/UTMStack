package config

import (
	"fmt"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbOnce sync.Once
	dbConn *gorm.DB
)

func InitDb() {
	dbOnce.Do(func() {
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", dbHost, dbPort, dbUser, dbPassword, dbName)
		var err error
		dbConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic(err)
		}
	})
}

func GetDB() *gorm.DB {
	if dbConn == nil {
		InitDb()
	}
	return dbConn
}
