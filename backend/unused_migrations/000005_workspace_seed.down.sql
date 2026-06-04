DELETE FROM workspaces WHERE slug = 'default' AND is_default = TRUE;

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('workspaces.read', 'workspaces.write', 'workspaces.delete')
);

DELETE FROM permissions WHERE name IN ('workspaces.read', 'workspaces.write', 'workspaces.delete');
