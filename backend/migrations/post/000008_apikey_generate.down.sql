DELETE FROM role_permission WHERE  permission_name = 'apikey.write';
DELETE FROM role_permission WHERE  permission_name = 'apikey.read';
DELETE FROM permissions WHERE name = 'apikey.write';
DELETE FROM permissions WHERE name = 'apikey.read';
