-- 000002_linux_user_auditor.up.sql
--
-- Purpose:
--   Extend the ad_user table so it can host both Windows (AD) users and Linux users
--   in a single inventory. Windows uniqueness stays on (tenant_id, sid); Linux
--   uniqueness moves to (tenant_id, machine_id, uid_number). Both are enforced by
--   Postgres partial unique indexes — the two universes cannot collide with each
--   other, and provisional Linux rows (machine_id IS NULL) coexist because Postgres
--   treats NULLs as distinct in unique indexes.
--
-- Ordering:
--   AutoMigrate (GORM) runs BEFORE this migration and adds the columns declared on
--   domain.ADUser (source, machine_id, uid_number, hostname, username) with their
--   defaults. This SQL file swaps the old full-table unique index for the two
--   source-discriminated partial indexes.
--
-- Safety:
--   ADD COLUMN IF NOT EXISTS is a belt-and-suspenders guard against environments
--   where AutoMigrate ran partially or was skipped.
--

-- Safety net: ensure the source column exists with the correct default before the
-- partial indexes reference it. AutoMigrate should have already added it, but this
-- is idempotent and covers partial-migration recovery paths.
ALTER TABLE ad_user
    ADD COLUMN IF NOT EXISTS source VARCHAR(16) NOT NULL DEFAULT 'windows';

-- Drop the old full-table unique index. Windows uniqueness moves to a partial
-- variant that only enforces on rows where source = 'windows'.
DROP INDEX IF EXISTS idx_ad_user_tenant_sid;

-- Windows partial unique index.
CREATE UNIQUE INDEX IF NOT EXISTS idx_aduser_windows
    ON ad_user (tenant_id, sid)
    WHERE source = 'windows';

-- Linux partial unique index. Rows where machine_id IS NULL are provisional and
-- Postgres treats their NULL as distinct — so multiple provisional rows for the
-- same (tenant_id, hostname, username) with machine_id = NULL are permitted at the
-- index level. Application-level de-duplication (see repository.Upsert) prevents
-- them in practice.
CREATE UNIQUE INDEX IF NOT EXISTS idx_aduser_linux
    ON ad_user (tenant_id, machine_id, uid_number)
    WHERE source = 'linux';
