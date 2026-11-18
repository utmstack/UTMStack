package upgrade

import (
	"database/sql"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// id5 derives a deterministic v12 UUID from a v11 numeric surrogate id under
// the entity's namespace. Re-runs always produce the same value.
func id5(ns uuid.UUID, v11ID int64) string {
	return uuid.NewSHA1(ns, []byte(strconv.FormatInt(v11ID, 10))).String()
}

// id5s is id5 for string keys (app_config uses the key itself, not a numeric id).
func id5s(ns uuid.UUID, s string) string {
	return uuid.NewSHA1(ns, []byte(s)).String()
}

// tableExists reports whether a table is present on the connected database.
// The upgrade tolerates an absent optional v11 table: a single converter
// skipping must not abort the whole conversion.
func tableExists(db *gorm.DB, name string) bool {
	var exists bool
	db.Raw(`SELECT to_regclass(?) IS NOT NULL`, name).Scan(&exists)
	return exists
}

func logSkipTable(table string) {
	catcher.Info("v11 upgrade: skipping missing table", map[string]any{"table": table})
}

func isV11FromFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			catcher.Warn("v11 upgrade: version-old.json unreadable", map[string]any{"path": path, "error": err.Error()})
		}
		return false
	}
	var v struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		catcher.Warn("v11 upgrade: version-old.json invalid", map[string]any{"path": path, "error": err.Error()})
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(v.Version), "v11.")
}

func connectAgentmanager(dsn string) (*sql.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return sqlDB, nil
}
