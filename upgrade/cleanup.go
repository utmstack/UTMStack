package upgrade

import (
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

// CleanupOrphans drops what a v11 conversion leaves behind in the utmstack
// database: tables only v11 created, sequences whose owning table was dropped,
// and views still built over v11 tables. It is a no-op on a base that was
// never a v11 install.
func CleanupOrphans(db *gorm.DB) error {
	if !isV11(db) {
		return nil
	}

	tables, err := v11TableList(db)
	if err != nil {
		return err
	}
	if err := dropEach(db, tables, "TABLE"); err != nil {
		return err
	}

	var seqs []string
	if err := db.Raw(`SELECT s.relname FROM pg_class s
		WHERE s.relkind = 'S'
		  AND s.relnamespace = 'public'::regnamespace
		  AND (s.relname LIKE 'utm\_%' ESCAPE '\'
		       OR s.relname LIKE 'jhi\_%' ESCAPE '\'
		       OR s.relname IN ('getting_started_seq', 'ira_variable_seq'))
		  AND NOT EXISTS (
			SELECT 1 FROM pg_depend d
			JOIN pg_class t ON t.oid = d.refobjid
			WHERE d.objid = s.oid
			  AND d.refobjid <> 0
			  AND t.relname IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public')
		  )`).Scan(&seqs).Error; err != nil {
		return err
	}
	if err := dropEach(db, seqs, "SEQUENCE"); err != nil {
		return err
	}

	var views []string
	if err := db.Raw(`SELECT v.relname FROM pg_class v
		JOIN pg_namespace n ON n.oid = v.relnamespace
		JOIN pg_depend d ON d.refobjid = v.oid AND d.deptype = 'n'
		JOIN pg_class c ON c.oid = d.objid
		WHERE n.nspname = 'public'
		  AND v.relkind = 'v'
		  AND (c.relname LIKE 'utm\_%' ESCAPE '\'
		       OR c.relname LIKE 'jhi\_%' ESCAPE '\'
		       OR c.relname IN ('api_keys_v11', 'databasechangelog', 'databasechangeloglock'))`).Scan(&views).Error; err != nil {
		return err
	}
	if err := dropEach(db, views, "VIEW"); err != nil {
		return err
	}

	catcher.Info("v11 upgrade: orphans cleaned", map[string]any{"tables": len(tables), "sequences": len(seqs), "views": len(views)})
	return nil
}

func v11TableList(db *gorm.DB) ([]string, error) {
	var tables []string
	if err := db.Raw(`SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
		  AND (tablename LIKE 'utm\_%' ESCAPE '\'
		       OR tablename LIKE 'jhi\_%' ESCAPE '\'
		       OR tablename IN ('api_keys_v11', 'databasechangelog', 'databasechangeloglock'))`).Scan(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

func dropEach(db *gorm.DB, names []string, kind string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, n := range names {
			if n == "" || strings.Contains(n, `"`) {
				continue
			}
			if err := tx.Exec(`DROP ` + kind + ` IF EXISTS public."` + n + `" CASCADE`).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
