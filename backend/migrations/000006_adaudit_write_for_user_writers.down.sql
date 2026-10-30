-- Reverse the 000006 grant. Cannot reconstruct who had users.write at
-- migration time, so strip adaudit.write from every role that has it
-- alongside users.write now.
DELETE FROM role_permission rp
WHERE rp.permission_name = 'adaudit.write'
  AND EXISTS (
      SELECT 1 FROM role_permission u
      WHERE u.role_id = rp.role_id
        AND u.permission_name = 'users.write'
  );
