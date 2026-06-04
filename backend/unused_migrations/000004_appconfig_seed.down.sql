DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('config.read', 'config.write')
);

DELETE FROM permissions WHERE name IN ('config.read', 'config.write');
