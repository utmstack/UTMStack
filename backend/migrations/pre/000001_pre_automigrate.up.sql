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

-- Compliance report schedules were reshaped (compliance_id/url_with_params/filters →
-- framework_key/recipients). The legacy rows referenced now-dropped standards, so they
-- are meaningless; drop the table here so AutoMigrate recreates it with the new schema.
DROP TABLE IF EXISTS public.utm_compliance_report_schedule CASCADE;
DROP SEQUENCE IF EXISTS public.utm_compliance_report_schedule_id_seq;

-- Dashboards were reshaped: visualizations are now SQL-only (sql_query + an opaque
-- config blob; the legacy query/aggregation/chart_* columns are gone). The existing
-- rows are meaningless under the new model and we don't need to preserve them, so
-- drop all three dashboard tables here; AutoMigrate recreates them with the new
-- schema. CASCADE handles the join's FKs.
DROP TABLE IF EXISTS public.utm_dashboard_visualization CASCADE;
DROP TABLE IF EXISTS public.utm_visualization CASCADE;
DROP TABLE IF EXISTS public.utm_dashboard CASCADE;
