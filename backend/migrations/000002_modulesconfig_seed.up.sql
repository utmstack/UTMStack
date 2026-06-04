-- Seed the modulesconfig RBAC permissions. The Go modulesconfig module gates
-- its REST surface on these two: reads (list/details/categories/requirements)
-- require modules.read; writes (activateDeactivate, group CRUD, config update)
-- require modules.write.
--
-- Follow the same pattern used for the seed permissions in 000001: insert into
-- the catalog with ON CONFLICT DO NOTHING (idempotent on re-runs), then bind
-- ROLE_ADMIN to both and ROLE_USER to the read.

INSERT INTO permissions (name, description, resource, action) VALUES
    ('modules.read',  'List and view application modules and their configuration', 'modules', 'read'),
    ('modules.write', 'Activate/deactivate modules and update module configuration', 'modules', 'write')
ON CONFLICT (name) DO NOTHING;

INSERT INTO authority_permissions (authority_name, permission_id)
SELECT 'ROLE_ADMIN', p.id
FROM permissions p
WHERE p.name IN ('modules.read', 'modules.write')
ON CONFLICT DO NOTHING;

INSERT INTO authority_permissions (authority_name, permission_id)
SELECT 'ROLE_USER', p.id
FROM permissions p
WHERE p.name = 'modules.read'
ON CONFLICT DO NOTHING;
