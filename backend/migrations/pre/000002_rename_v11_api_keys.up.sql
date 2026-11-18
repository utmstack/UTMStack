-- A v11 base carries an api_keys table with a different shape (bigint id,
-- plaintext key) than v12 (uuid id, key_hash). Rename the v11 copy so the
-- converter can read it and GORM AutoMigrate creates a fresh v12 api_keys
-- without colliding with the old schema.
DO $$
BEGIN
    IF to_regclass('jhi_user') IS NOT NULL
       AND to_regclass('api_keys') IS NOT NULL
       AND to_regclass('api_keys_v11') IS NULL THEN
        ALTER TABLE api_keys RENAME TO api_keys_v11;
    END IF;
END $$;
