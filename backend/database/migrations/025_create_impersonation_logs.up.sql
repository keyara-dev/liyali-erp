-- Migration: Create impersonation_logs table
-- Tracks all impersonation events for audit and security purposes.
-- Only super_admin users can view this log via the admin console.

CREATE TABLE IF NOT EXISTS impersonation_logs (
    id TEXT PRIMARY KEY,
    impersonator_id TEXT NOT NULL,
    impersonator_email TEXT NOT NULL,
    target_id TEXT NOT NULL,
    target_email TEXT NOT NULL,
    -- platform_user: admin impersonating a regular platform user
    -- admin_user: super_admin impersonating another admin/super_admin user
    impersonation_type TEXT NOT NULL CHECK (impersonation_type IN ('platform_user', 'admin_user')),
    token_jti TEXT NOT NULL,
    reason TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT false,
    revoked_at TIMESTAMPTZ,
    revoked_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_impersonation_logs_impersonator ON impersonation_logs(impersonator_id);
CREATE INDEX IF NOT EXISTS idx_impersonation_logs_target ON impersonation_logs(target_id);
CREATE INDEX IF NOT EXISTS idx_impersonation_logs_created_at ON impersonation_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_impersonation_logs_type ON impersonation_logs(impersonation_type);
