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

-- SOAR rule executions are now keyed by the flow's file identity (rule_path) instead
-- of the legacy numeric rule_id FK into utm_alert_response_rule (which the soar flow
-- bootstrap drops once flows migrate to YAML). The old not-null rule_id column would
-- block the new path-based inserts, and executions are transient operational records
-- (no curation to preserve), so drop the table here; AutoMigrate recreates it with the
-- rule_path schema.
DROP TABLE IF EXISTS public.utm_alert_response_rule_execution CASCADE;
DROP SEQUENCE IF EXISTS public.utm_alert_response_rule_execution_id_seq;

-- Dashboards were reshaped: visualizations are now SQL-only (sql_query + an opaque
-- config blob; the legacy query/aggregation/chart_* columns are gone). The existing
-- rows are meaningless under the new model and we don't need to preserve them, so
-- drop all three dashboard tables here; AutoMigrate recreates them with the new
-- schema. CASCADE handles the join's FKs.
DROP TABLE IF EXISTS public.utm_dashboard_visualization CASCADE;
DROP TABLE IF EXISTS public.utm_visualization CASCADE;
DROP TABLE IF EXISTS public.utm_dashboard CASCADE;

-- Legacy reporting view (joins utm_server/utm_module/utm_module_group/…). It is a
-- relic of the Java backend — the Go backend never queries it — and it pins the
-- types of the columns it selects, so AutoMigrate fails to resize utm_module
-- columns ("cannot alter type of a column used by a view or rule"). Drop it.
DROP VIEW IF EXISTS public.utm_server_configurations CASCADE;

-- Reconcile legacy (JHipster/liquibase) single-column UNIQUE CONSTRAINTS with the
-- Go models, which declare these columns as `uniqueIndex` (a unique INDEX, not a
-- named column constraint). GORM AutoMigrate sees the column-level unique
-- constraint on the existing table and tries to drop it under the name IT would
-- have used (uni_<table>_<column>), which fails because the legacy constraint has a
-- different name (e.g. ux_user_login):
--   ERROR: constraint "uni_jhi_user_login" of relation "jhi_user" does not exist
--
-- Drop the legacy single-column unique constraints here — looked up by table+column
-- so we don't depend on the (instance-specific) constraint name — and AutoMigrate
-- then recreates each as the unique index it expects. Uniqueness is preserved (the
-- data already satisfies it; the index just gets a new name, idx_<table>_<column>).
-- Idempotent: only acts on constraints that exist, so it's a no-op on a fresh
-- install or when already reconciled.
DO $$
DECLARE
    r record;
BEGIN
    FOR r IN
        SELECT c.conrelid::regclass::text AS tbl, c.conname
        FROM pg_constraint c
        JOIN pg_attribute a
          ON a.attrelid = c.conrelid AND a.attnum = c.conkey[1]
        WHERE c.contype = 'u'
          AND cardinality(c.conkey) = 1            -- single-column constraints only
          AND (c.conrelid::regclass::text, a.attname) IN (
                VALUES
                    ('jhi_user', 'login'),
                    ('jhi_user', 'email'),
                    ('utm_alert_tag', 'tag_name'),
                    ('utm_alert_tag_rule', 'rule_name'),
                    ('utm_incident', 'incident_name'),
                    ('utm_incident_alert', 'alert_id'),
                    ('utm_index_pattern', 'pattern'),
                    ('utm_incident_variables', 'variable_name')
          )
    LOOP
        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT %I', r.tbl, r.conname);
        RAISE NOTICE 'reconcile: dropped legacy unique constraint % on %', r.conname, r.tbl;
    END LOOP;
END $$;

-- utm_asset_group: the legacy schema had a `type` discriminator (ASSET/COLLECTOR,
-- liquibase 20240506001) and a composite unique over (group_name, type). The Go
-- model (UtmAssetGroup) drops `type` and makes group_name globally unique via a
-- uniqueIndex. AutoMigrate creates that single-column unique index, and it runs
-- BEFORE the post-AutoMigrate SQL — so if two rows share a group_name across an
-- ASSET and a COLLECTOR group, "CREATE UNIQUE INDEX ... (group_name)" fails. The
-- reconciliation must therefore happen HERE, in pre.
--
-- Nothing references utm_asset_group.id at this point: the only FK to it is the new
-- `datasources.group_id`, and `datasources` is created later by AutoMigrate and
-- populated at runtime (group linkage resolved by name). So duplicate group_names
-- can simply be collapsed to the lowest id — no repointing needed.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'utm_asset_group' AND column_name = 'type'
    ) THEN
        DELETE FROM utm_asset_group g
        USING (
            SELECT group_name, min(id) AS keep_id
            FROM utm_asset_group GROUP BY group_name
        ) k
        WHERE g.group_name = k.group_name AND g.id <> k.keep_id;

        -- CASCADE removes the legacy composite unique constraint (whatever its name)
        -- that included `type`; AutoMigrate then builds idx_utm_asset_group_group_name.
        ALTER TABLE utm_asset_group DROP COLUMN type CASCADE;
    END IF;
END $$;
