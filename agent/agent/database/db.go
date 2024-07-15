package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbInstance *Database
	dbOnce     sync.Once
)

type Database struct {
	db     *gorm.DB
	locker sync.RWMutex
}

func (d *Database) Migrate(data interface{}) error {
	return d.db.AutoMigrate(data)
}

func (d *Database) Create(data interface{}) error {
	return d.db.Create(data).Error
}

func (d *Database) Find(data interface{}, field string, value interface{}) (bool, error) {
	err := d.db.Where(fmt.Sprintf("%v = ?", field), value).Find(data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *Database) GetAll(data interface{}) error {
	if err := d.db.Find(data).Error; err != nil {
		return err
	}
	return nil
}

func (d *Database) Update(data interface{}, searchField string, searchValue string, modifyField string, newValue interface{}) error {
	return d.db.Model(data).Where(fmt.Sprintf("%v = ?", searchField), searchValue).Update(modifyField, newValue).Error
}

func (d *Database) Delete(data interface{}, field string, value string) error {
	return d.db.Where(fmt.Sprintf("%v = ?", field), value).Delete(data).Error
}

func (d *Database) DeleteOld(data interface{}, retentionMinutes int) (int64, error) {
	cutoffDateTime := time.Now().Add(-time.Minute * time.Duration(retentionMinutes))
	cutoffDateTimeFormatted := cutoffDateTime.Format("2006-01-02 15:04:05")
	result := d.db.Where("created_at < ?", cutoffDateTimeFormatted).Delete(data)
	if result.Error != nil {
		return 0, result.Error
	}

	err := d.db.Exec("VACUUM;").Error
	if err != nil {
		return result.RowsAffected, err
	}

	return result.RowsAffected, nil
}

func (d *Database) Lock() {
	d.locker.Lock()
}

func (d *Database) Unlock() {
	d.locker.Unlock()
}

func GetDB() *Database {
	dbOnce.Do(func() {
		path := filepath.Join(utils.GetMyPath(), "logs_process")
		err := utils.CreatePathIfNotExist(path)
		if err != nil {
			log.Fatalf("error creating database path: %v", err)
		}
		path = filepath.Join(path, config.LogsDBFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			file, err := os.Create(path)
			if err != nil {
				log.Fatalf("error creating database file: %v", err)
			}
			file.Close()
		}

		conn, err := gorm.Open(sqlite.Open(path+"?cache=shared"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Fatalf("error connecting with database: %v", err)
		}

		dbInstance = &Database{db: conn}

	})

	return dbInstance
}
