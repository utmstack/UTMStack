-- Pre-AutoMigrate migration: runs BEFORE GORM AutoMigrate (tracked in its own
-- schema_migrations_pre table, so it runs exactly once).
--
-- The integrations model adds data_type (NOT NULL + UNIQUE) and is_system to
-- utm_module. AutoMigrate cannot add a NOT NULL column to an already-populated
-- table, so on an in-place upgrade we add + backfill them here; AutoMigrate then
-- only enforces the constraints (now satisfiable). Guarded: no-op on a fresh
-- install where utm_module does not exist yet.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'utm_module') THEN
        ALTER TABLE utm_module ADD COLUMN IF NOT EXISTS data_type varchar(250);
        ALTER TABLE utm_module ADD COLUMN IF NOT EXISTS is_system boolean NOT NULL DEFAULT false;

        -- Backfill data_type from the legacy module<->dataType mapping.
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'utm_data_types') THEN
            UPDATE utm_module m
            SET data_type = dt.data_type
            FROM utm_data_types dt
            WHERE dt.module_id = m.id AND (m.data_type IS NULL OR m.data_type = '');
        END IF;

        -- Modules with no dataType mapping (oracle, and the non-source modules
        -- like soc_ai/file_integrity): derive from module_name so NOT NULL holds
        -- and stays unique (module_name is unique).
        UPDATE utm_module SET data_type = lower(module_name)
        WHERE data_type IS NULL OR data_type = '';

        -- Everything that existed before this release is a built-in integration.
        UPDATE utm_module SET is_system = true WHERE is_system = false;
    END IF;
END $$;
