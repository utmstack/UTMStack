-- Workspace: register the permissions that gate the CRUD endpoints, bind
-- them to the admin role, and seed a single "Default" workspace so a fresh
-- install is never empty. The default workspace owns every existing record
-- when more workspaces don't yet exist.

INSERT INTO permissions (name, description, resource, action) VALUES
    ('workspaces.read',   'List and view workspaces',   'workspaces', 'read'),
    ('workspaces.write',  'Create and update workspaces', 'workspaces', 'write'),
    ('workspaces.delete', 'Delete workspaces',          'workspaces', 'delete')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN ('workspaces.read', 'workspaces.write', 'workspaces.delete')
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO workspaces (slug, name, description, is_default, created_at, updated_at, created_by)
SELECT 'default', 'Default', 'Default workspace created at install time.', TRUE, NOW(), NOW(), 'system'
WHERE NOT EXISTS (SELECT 1 FROM workspaces);
