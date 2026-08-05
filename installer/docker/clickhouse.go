package docker

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

// clickHouseSchema is the table definition, embedded so an install carries it
// rather than fetching it. The server runs whatever is in its init directory on
// first boot, which is what creates the tables — nothing else does.
//
// It lives in this module because go:embed cannot reach outside it, and one
// copy that the installer ships beats two that drift.
//
//go:embed clickhouse-schema.sql
var clickHouseSchema []byte

const (
	schemaFileName   = "01-schema.sql"
	settingsFileName = "utmstack-settings.xml"
)

//go:embed clickhouse-settings.xml
var clickHouseSettings []byte

// writeClickHouseSettings pins the server settings the pipeline depends on.
// Unlike the schema, this is read on every start, so it applies to an existing
// install as well as a new one.
func writeClickHouseSettings(dir string) error {
	path := filepath.Join(dir, settingsFileName)
	if err := os.WriteFile(path, clickHouseSettings, 0o644); err != nil {
		return fmt.Errorf("writing the ClickHouse settings to %s: %w", path, err)
	}
	return nil
}

// writeClickHouseSchema puts the schema where the server will find it. Writing
// it on every run keeps an upgraded installer's schema from being shadowed by
// an older one; the statements are CREATE TABLE IF NOT EXISTS, and the server
// only runs them when the data directory is empty.
func writeClickHouseSchema(dir string) error {
	path := filepath.Join(dir, schemaFileName)
	if err := os.WriteFile(path, clickHouseSchema, 0o644); err != nil {
		return fmt.Errorf("writing the ClickHouse schema to %s: %w", path, err)
	}
	return nil
}
