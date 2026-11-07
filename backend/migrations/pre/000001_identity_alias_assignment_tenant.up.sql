-- Ensure the column exists before backfilling; idempotent.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'identity_alias_assignment'
          AND column_name = 'tenant_id'
    ) THEN
        ALTER TABLE identity_alias_assignment
            ADD COLUMN tenant_id uuid;
    END IF;
END $$;

UPDATE identity_alias_assignment assignment
SET tenant_id = user_row.tenant_id::uuid
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
