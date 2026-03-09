-- Create user_activity_logs table for tracking all user actions
-- Note: No FK constraint to avoid lock contention on users table
CREATE TABLE IF NOT EXISTS user_activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255),
    action_type VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ual_user_id ON user_activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_ual_created_at ON user_activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ual_action_type ON user_activity_logs(action_type);
CREATE INDEX IF NOT EXISTS idx_ual_user_created ON user_activity_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ual_org_created ON user_activity_logs(organization_id, created_at DESC);
