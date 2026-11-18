DROP INDEX IF EXISTS idx_identity_alias_assignment_tenant;

DO $$
BEGIN
    IF to_regclass('identity_alias_assignment') IS NOT NULL THEN
        ALTER TABLE identity_alias_assignment
            DROP COLUMN IF EXISTS tenant_id;
    END IF;
END $$;
