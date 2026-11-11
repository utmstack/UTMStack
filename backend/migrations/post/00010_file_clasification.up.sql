


INSERT INTO integrations (tenant_id, name, data_type, ingest_type, system_owner) VALUES
    ('00000000-0000-0000-0000-000000000000', 'FILE_CLASSIFICATION',  'wineventlog', 'agent', TRUE)
ON CONFLICT (tenant_id, name) DO UPDATE SET
    data_type   = EXCLUDED.data_type,
    ingest_type = EXCLUDED.ingest_type;
