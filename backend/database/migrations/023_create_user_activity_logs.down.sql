-- Drop user_activity_logs table and its indexes
DROP INDEX IF EXISTS idx_ual_metadata;
DROP INDEX IF EXISTS idx_ual_org_created;
DROP INDEX IF EXISTS idx_ual_user_created;
DROP INDEX IF EXISTS idx_ual_action_type;
DROP INDEX IF EXISTS idx_ual_created_at;
DROP INDEX IF EXISTS idx_ual_user_id;
DROP TABLE IF EXISTS user_activity_logs;
