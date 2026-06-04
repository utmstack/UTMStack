-- Reverse 000002_modulesconfig_seed: drop the role bindings and the permissions
-- themselves. GORM AutoMigrate handles the table schema, so this only undoes
-- the seeded rows.

DELETE FROM authority_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('modules.read', 'modules.write')
);

DELETE FROM permissions WHERE name IN ('modules.read', 'modules.write');
