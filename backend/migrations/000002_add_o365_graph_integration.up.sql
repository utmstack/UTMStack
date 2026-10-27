-- v12.0.0 was released with the integrations catalog seeded in 000001 alone.
-- O365_GRAPH arrived after that release, so existing installs already ran
-- 000001 and never see the row; golang-migrate only applies new versions on
-- upgrade, so the row is added here instead. Idempotent on both fresh and
-- upgraded databases.
INSERT INTO integrations (tenant_id, name, data_type, ingest_type, system_owner)
VALUES ('00000000-0000-0000-0000-000000000000', 'O365_GRAPH', 'o365-graph', 'plugin', TRUE)
ON CONFLICT (tenant_id, name) DO NOTHING;
