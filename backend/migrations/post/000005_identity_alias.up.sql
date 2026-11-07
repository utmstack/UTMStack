-- The identity alias tables are created by GORM AutoMigrate (see
-- database.Models), which cannot express two things this feature needs:
--  * names are unique case-insensitively per tenant -> functional index,
--  * assignments cascade from both ends -> FK constraints.
-- Idempotent on both fresh and upgraded databases.
CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_alias_tenant_name
    ON identity_alias (tenant_id, lower(name));

ALTER TABLE identity_alias_assignment
    DROP CONSTRAINT IF EXISTS fk_identity_alias_assignment_alias,
    DROP CONSTRAINT IF EXISTS fk_identity_alias_assignment_user,
    ADD CONSTRAINT fk_identity_alias_assignment_alias
        FOREIGN KEY (alias_id) REFERENCES identity_alias (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_identity_alias_assignment_user
        FOREIGN KEY (user_id) REFERENCES ad_user (id) ON DELETE CASCADE;
