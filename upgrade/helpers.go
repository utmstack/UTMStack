package upgrade

import (
	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

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
