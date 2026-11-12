DO $$
BEGIN
    IF to_regclass('api_keys_v11') IS NOT NULL
       AND to_regclass('api_keys') IS NULL THEN
        ALTER TABLE api_keys_v11 RENAME TO api_keys;
    END IF;
END $$;
