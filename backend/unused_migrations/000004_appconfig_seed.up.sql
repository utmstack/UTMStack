-- App-config: register the permissions that gate the CRUD endpoints and
-- bind them to the admin role. Each settings page (connection key, email,
-- data retention, etc.) reads/writes its keys via /api/v1/config — there
-- are no per-feature permissions, just the generic config.read/write.

INSERT INTO permissions (name, description, resource, action) VALUES
    ('config.read',  'Read application config',             'config', 'read'),
    ('config.write', 'Update or delete application config', 'config', 'write')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('config.read', 'config.write')
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;
