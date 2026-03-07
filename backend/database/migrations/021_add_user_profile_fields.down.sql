-- Remove indexes
DROP INDEX IF EXISTS idx_users_nrc_number;
DROP INDEX IF EXISTS idx_users_man_number;

-- Remove columns
ALTER TABLE users DROP COLUMN IF EXISTS contact;
ALTER TABLE users DROP COLUMN IF EXISTS nrc_number;
ALTER TABLE users DROP COLUMN IF EXISTS man_number;
ALTER TABLE users DROP COLUMN IF EXISTS position;
