-- The up migration only seeds content the product ships. Removing it would
-- leave an install unable to start — no roles, no permissions, no configuration
-- — so this is deliberately empty rather than a mirror of the seeds.
--
-- There is nothing to undo: v12 is installed clean, so a rollback is a fresh
-- install, not a downgrade of an existing one.

SELECT 1;
