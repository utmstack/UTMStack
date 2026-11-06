DROP INDEX IF EXISTS idx_identity_alias_assignment_tenant;

ALTER TABLE identity_alias_assignment
    DROP COLUMN IF EXISTS tenant_id;
