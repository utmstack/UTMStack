-- changeset author:update-windows-agent-pattern
UPDATE utm_user_source
SET index_pattern = 'v11-log-wineventlog-*',
    modified_date = now()
WHERE id = 1 AND index_name = 'WINDOWS_AGENT';


