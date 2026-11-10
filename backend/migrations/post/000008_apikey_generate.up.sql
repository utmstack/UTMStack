-- Identity aliases (operator-named, one link per account) need a write
-- permission. The init migration seeded adaudit.read only, and 000001 has
-- already run on every live install, so the new row lands here instead.
-- Idempotent on both fresh and upgraded databases.
INSERT INTO permissions (name, description) VALUES
    ('apikey.write', 'Generate api keys'),
    ('apikey.read', 'Make an apikey list')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permission (role_id, permission_name)
SELECT '00000000-0000-0000-0000-00000000a001', 'apikey.write'
ON CONFLICT DO NOTHING;


INSERT INTO role_permission (role_id, permission_name)
SELECT '00000000-0000-0000-0000-00000000a002', 'apikey.read'
ON CONFLICT DO NOTHING;

INSERT INTO role_permission (role_id, permission_name)
SELECT '00000000-0000-0000-0000-00000000a003', 'apikey.read'
ON CONFLICT DO NOTHING;

