-- 000002_linux_user_auditor.down.sql
--
-- Reverse the partial-index topology and restore the full-table unique index.
-- The additive columns (source, machine_id, uid_number, hostname, username) are
-- intentionally NOT dropped — leaving them in place is harmless for the old code
-- path and preserves data. If a full schema rollback is required, drop the
-- columns manually after this migration runs.
--
-- Ordering caveat (documented in the runbook):
--   Rolling back the backend code without running this .down.sql leaves the old
--   Windows ON CONFLICT (tenant_id, sid) statement targeting an index that no
--   longer exists. The Windows upsert path will fail. ALWAYS run this .down.sql
--   BEFORE rolling back the code.

DROP INDEX IF EXISTS idx_aduser_linux;
DROP INDEX IF EXISTS idx_aduser_windows;

CREATE UNIQUE INDEX IF NOT EXISTS idx_ad_user_tenant_sid
    ON ad_user (tenant_id, sid);
