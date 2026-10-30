-- Users who can write users (users.write) should also be able to write
-- identity aliases (adaudit.write) — same trust level. Grants the permission
-- to every role that holds users.write, so the effect is on all users
-- assigned those roles at migration time. Idempotent.
INSERT INTO role_permission (role_id, permission_name)
SELECT r.role_id, 'adaudit.write'
FROM role_permission r
WHERE r.permission_name = 'users.write'
ON CONFLICT DO NOTHING;
