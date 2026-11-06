DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM identity_alias_assignment assignment
        JOIN ad_user user_row ON user_row.id = assignment.user_id
        JOIN identity_alias alias_row ON alias_row.id = assignment.alias_id
        WHERE user_row.tenant_id <> alias_row.tenant_id
    ) THEN
        RAISE EXCEPTION 'identity alias assignments cross tenant boundaries; resolve them before migration';
    END IF;
END $$;

UPDATE identity_alias_assignment assignment
SET tenant_id = user_row.tenant_id
FROM ad_user user_row, identity_alias alias_row
WHERE assignment.user_id = user_row.id
  AND assignment.alias_id = alias_row.id
  AND user_row.tenant_id = alias_row.tenant_id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM identity_alias_assignment
        WHERE tenant_id IS NULL
    ) THEN
        RAISE EXCEPTION 'identity alias assignments could not be backfilled with a tenant';
    END IF;
END $$;

ALTER TABLE identity_alias_assignment
    ALTER COLUMN tenant_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_identity_alias_assignment_tenant
    ON identity_alias_assignment (tenant_id);
