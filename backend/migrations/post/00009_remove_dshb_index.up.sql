DROP INDEX IF EXISTS idx_integration_tenant_data_type;
ALTER TABLE integrations
    DROP CONSTRAINT IF EXISTS idx_integration_tenant_data_type;
