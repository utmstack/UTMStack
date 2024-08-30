package database

import (
	"errors"
	"fmt"
	"sync"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbOnce     sync.Once
	dbInstance *DB
)

type DB struct {
	conn   *gorm.DB
	locker sync.RWMutex
}

func (d *DB) Migrate(data ...interface{}) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	for _, v := range data {
		if err := d.conn.AutoMigrate(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) Create(data interface{}) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	if err := d.conn.Create(data).Error; err != nil {
		return err
	}
	return nil
}

func (d *DB) Upsert(data interface{}, query string, updates map[string]interface{}, args ...interface{}) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	var count int64
	if err := d.conn.Model(data).Where(query, args...).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		if updates != nil {
			return d.conn.Model(data).Where(query, args...).Updates(updates).Error
		}
		return d.conn.Model(data).Where(query, args...).Updates(data).Error
	}
	return d.conn.Create(data).Error
}

func (d *DB) GetFirst(data interface{}, query string, args ...interface{}) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	err := d.conn.Where(query, args).First(data).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) GetAll(data interface{}, query string, args ...interface{}) (int64, error) {
	d.locker.Lock()
	defer d.locker.Unlock()
	tx := d.conn
	if query != "" {
		tx = tx.Where(query, args...)
	}
	result := tx.Find(data)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (d *DB) GetByPagination(data interface{}, p utils.Pagination, f []utils.Filter, join string, getDeleted bool) (int64, error) {
	d.locker.Lock()
	defer d.locker.Unlock()
	var count int64
	tx := d.conn.Model(data).Scopes(utils.FilterScope(f)).Count(&count).Scopes(p.PagingScope)
	if getDeleted {
		tx = tx.Unscoped()
	}
	if join != "" {
		tx = tx.Joins(join)
	}
	tx = tx.Find(data)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return count, nil
}

func (d *DB) Delete(data interface{}, query string, hardDelete bool, args ...interface{}) error {
	d.locker.Lock()
	defer d.locker.Unlock()
	tx := d.conn
	if hardDelete {
		tx = tx.Unscoped()
	}
	err := tx.Where(query, args).Delete(data).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func GetDB() *DB {
	dbOnce.Do(func() {
		dbInstance = &DB{}
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", config.GetDBHost(), config.GetDBPort(), config.GetDBUser(), config.GetDBPassword(), config.GetDBName())
		var err error
		dbInstance.conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic(err)
		}
	})
	return dbInstance
}
