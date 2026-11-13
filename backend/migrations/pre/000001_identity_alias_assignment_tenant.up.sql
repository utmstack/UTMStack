DO $$
BEGIN
    -- Absent on a v11 base; AutoMigrate creates it fully with tenant_id, so there is
    -- nothing to backfill and every statement below would fail on the missing table.
    IF to_regclass('identity_alias_assignment') IS NULL THEN
        RETURN;
    END IF;

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

    UPDATE identity_alias_assignment assignment
    SET tenant_id = user_row.tenant_id::uuid
    FROM ad_user user_row, identity_alias alias_row
    WHERE assignment.user_id = user_row.id
      AND assignment.alias_id = alias_row.id
      AND user_row.tenant_id = alias_row.tenant_id;

    IF EXISTS (
        SELECT 1
        FROM identity_alias_assignment
        WHERE tenant_id IS NULL
    ) THEN
        RAISE EXCEPTION 'identity alias assignments could not be backfilled with a tenant';
    END IF;

    ALTER TABLE identity_alias_assignment
        ALTER COLUMN tenant_id SET NOT NULL;

    CREATE INDEX IF NOT EXISTS idx_identity_alias_assignment_tenant
        ON identity_alias_assignment (tenant_id);
END $$;
