DROP INDEX IF EXISTS idx_identity_alias_tenant_name;
ALTER TABLE identity_alias_assignment
    DROP CONSTRAINT IF EXISTS fk_identity_alias_assignment_user,
    DROP CONSTRAINT IF EXISTS fk_identity_alias_assignment_alias;
DROP TABLE IF EXISTS identity_alias_assignment;
DROP TABLE IF EXISTS identity_alias;
