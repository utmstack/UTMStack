package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/utils"
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

func (d *Database) DeleteOld(data interface{}, retentionMegabytes int) (int, error) {
	currentSize, err := GetDatabaseSizeInMB()
	if err != nil {
		return 0, fmt.Errorf("error getting database size: %v", err)
	}

	var rowsAffected int
	for currentSize > retentionMegabytes {
		result := d.db.Where("1 = 1").Order("created_at ASC").Limit(10).Delete(data)
		if result.Error != nil {
			return rowsAffected, result.Error
		}
		rowsAffected += int(result.RowsAffected)
		d.db.Exec("VACUUM;")
		currentSize, err = GetDatabaseSizeInMB()
		if err != nil {
			return rowsAffected, fmt.Errorf("error getting database size: %v", err)
		}
	}

	return rowsAffected, nil
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

		conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Fatalf("error connecting with database: %v", err)
		}

		dbInstance = &Database{db: conn}

	})

	return dbInstance
}

func GetDatabaseSizeInMB() (int, error) {
	path := filepath.Join(utils.GetMyPath(), "logs_process", config.LogsDBFile)
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return int(fileInfo.Size() / (1024 * 1024)), nil
}
