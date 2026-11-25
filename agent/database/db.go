package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/shared/fs"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbInstance *Database
	dbOnce     sync.Once
	dbInitErr  error
)

const (
	deleteBatchSize = 500
	insertBatchSize = 100
)

type Database struct {
	db     *gorm.DB
	locker sync.RWMutex
}

func (d *Database) Migrate(data interface{}) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	return d.db.AutoMigrate(data)
}

func (d *Database) Create(data interface{}) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	return d.db.Create(data).Error
}

func (d *Database) CreateBatch(logs []*models.Log) error {
	if len(logs) == 0 {
		return nil
	}
	d.locker.Lock()
	defer d.locker.Unlock()
	return d.db.CreateInBatches(logs, insertBatchSize).Error
}

func (d *Database) MarkProcessedBatch(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	d.locker.Lock()
	defer d.locker.Unlock()
	return d.db.Model(&models.Log{}).Where("id IN ?", ids).Update("processed", true).Error
}

func (d *Database) Find(data interface{}, field string, value interface{}) (bool, error) {
	d.locker.RLock()
	defer d.locker.RUnlock()
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
	d.locker.RLock()
	defer d.locker.RUnlock()
	if err := d.db.Find(data).Error; err != nil {
		return err
	}
	return nil
}

func (d *Database) Update(data interface{}, searchField string, searchValue string, modifyField string, newValue interface{}) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	return d.db.Model(data).Where(fmt.Sprintf("%v = ?", searchField), searchValue).Update(modifyField, newValue).Error
}

func (d *Database) Delete(data interface{}, field string, value string) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	return d.db.Where(fmt.Sprintf("%v = ?", field), value).Delete(data).Error
}

func (d *Database) Close() error {
	d.locker.Lock()
	defer d.locker.Unlock()
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

const HardCapMultiplier = 5

func (d *Database) DeleteOld(data interface{}, retentionMegabytes int) (processedDeleted int, unprocessedDeleted int, err error) {
	d.locker.Lock()
	defer d.locker.Unlock()
	currentSize, err := GetDatabaseSizeInMB()
	if err != nil {
		return 0, 0, fmt.Errorf("error getting database size: %v", err)
	}

	for currentSize > retentionMegabytes {
		result := d.deleteOldestBatch(data, "processed = ?", true)
		if result.Error != nil || result.RowsAffected == 0 {
			break
		}
		processedDeleted += int(result.RowsAffected)
		currentSize, err = GetDatabaseSizeInMB()
		if err != nil {
			break
		}
	}

	hardCap := retentionMegabytes * HardCapMultiplier
	for currentSize > hardCap {
		result := d.deleteOldestBatch(data, "processed = ?", false)
		if result.Error != nil || result.RowsAffected == 0 {
			break
		}
		unprocessedDeleted += int(result.RowsAffected)
		currentSize, err = GetDatabaseSizeInMB()
		if err != nil {
			break
		}
	}

	if processedDeleted+unprocessedDeleted > 0 {
		d.db.Exec("VACUUM;")
	}

	return processedDeleted, unprocessedDeleted, nil
}

func (d *Database) deleteOldestBatch(data interface{}, cond string, condArgs ...interface{}) *gorm.DB {
	subquery := d.db.Model(data).Select("id").
		Where(cond, condArgs...).
		Order("created_at ASC").
		Limit(deleteBatchSize)
	return d.db.Where("id IN (?)", subquery).Delete(data)
}

// Open initializes a SQLite database at path with WAL mode and busy timeout
// configured for concurrent high-throughput log buffering.
func Open(path string) (*Database, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("creating database file: %w", err)
		}
		file.Close()
	}

	conn, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("connecting with database: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA busy_timeout=5000;",
	} {
		if err := conn.Exec(pragma).Error; err != nil {
			return nil, fmt.Errorf("applying sqlite pragma %q: %w", pragma, err)
		}
	}

	return &Database{db: conn}, nil
}

func GetDB() (*Database, error) {
	dbOnce.Do(func() {
		path := filepath.Join(fs.GetExecutablePath(), "logs_process")
		if err := fs.CreateDirIfNotExist(path); err != nil {
			dbInitErr = fmt.Errorf("creating database path: %w", err)
			return
		}

		dbInstance, dbInitErr = Open(config.LogsDBFile)
	})

	if dbInitErr != nil {
		return nil, dbInitErr
	}
	return dbInstance, nil
}

func GetDatabaseSizeInMB() (int, error) {
	fileInfo, err := os.Stat(config.LogsDBFile)
	if err != nil {
		return 0, err
	}
	totalSize := fileInfo.Size()
	if walInfo, err := os.Stat(config.LogsDBFile + "-wal"); err == nil {
		totalSize += walInfo.Size()
	}
	return int(totalSize / (1024 * 1024)), nil
}
