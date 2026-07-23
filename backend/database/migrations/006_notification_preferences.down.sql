-- Migration: Drop notification_preferences table
-- Version: 006
-- Description: Drops the notification_preferences table

-- Drop trigger first
DROP TRIGGER IF EXISTS notification_preferences_updated_at_trigger ON notification_preferences;

-- Drop function
DROP FUNCTION IF EXISTS update_notification_preferences_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_notification_preferences_user_org;
DROP INDEX IF EXISTS idx_notification_preferences_org_id;
DROP INDEX IF EXISTS idx_notification_preferences_user_id;

-- Drop table
DROP TABLE IF EXISTS notification_preferences;