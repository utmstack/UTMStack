-- Manual flow trigger permission. Deliberately admin-only: it runs a flow
-- without alert matching, so it must not sit with soar.write (analysts have
-- that). Idempotent.
INSERT INTO permissions (name, description) VALUES
    ('soar.execute', 'Manually trigger SOAR response flows')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permission (role_id, permission_name)
VALUES ('00000000-0000-0000-0000-00000000a001', 'soar.execute')
ON CONFLICT DO NOTHING;
