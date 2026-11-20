package upgrade

import (
	"fmt"
	"os"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

// cleanupOrphans drops what a v11 conversion leaves behind in the utmstack
// database: tables only v11 created, sequences whose owning table was dropped,
// and views still built over v11 tables. Run calls it only after every
// converter succeeded in the same invocation, so it never deletes v11 source
// tables that were not converted. It also removes the v11 YAML files the old
// event processor left at the rules/filters roots; v12 only reads the
// system/ and user/ subdirectories.
func cleanupOrphans(db *gorm.DB, cnf Config) error {
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

	removed, err := removeV11Files(cnf.RulesDir, cnf.FiltersDir)
	if err != nil {
		return err
	}

	catcher.Info("v11 upgrade: rules and filters cleaned", map[string]any{"files": removed})
	catcher.Info("v11 upgrade: orphans cleaned", map[string]any{"tables": len(tables), "sequences": len(seqs), "views": len(views)})
	return nil
}

// removeV11Files deletes the numeric-named <id>.yaml files the v11 event
// processor wrote at the rules and filters roots. v12 never creates a file at
// those roots (the stores only write under system/ and user/), so the
// numeric-name pattern uniquely identifies v11 leftovers; subdirectories are
// never touched.
func removeV11Files(dirs ...string) (int, error) {
	removed := 0
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return removed, fmt.Errorf("read %s: %w", dir, err)
		}
		for _, e := range entries {
			if !e.Type().IsRegular() || !v11FileName(e.Name()) {
				continue
			}
			if err := os.Remove(dir + string(os.PathSeparator) + e.Name()); err != nil {
				return removed, fmt.Errorf("remove %s/%s: %w", dir, e.Name(), err)
			}
			removed++
		}
	}
	return removed, nil
}

func v11FileName(name string) bool {
	if len(name) <= 5 || !strings.HasSuffix(name, ".yaml") {
		return false
	}
	for _, c := range name[:len(name)-len(".yaml")] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
