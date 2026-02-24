-- Migration Rollback: 014_subscription_tier_system
-- Description: Rollback 3-tier system to previous 4-tier system
-- Date: 2026-02-24

-- ============================================================================
-- STEP 1: Restore old tier names in organizations
-- ============================================================================

-- Map new tier names back to old tier names
UPDATE organizations
SET subscription_tier = CASE subscription_tier
    WHEN 'starter' THEN 'basic'
    WHEN 'pro' THEN 'professional'
    WHEN 'custom' THEN 'enterprise'
    ELSE subscription_tier
END
WHERE subscription_tier IN ('starter', 'pro', 'custom');

-- ============================================================================
-- STEP 2: Restore old subscription tiers
-- ============================================================================

-- Delete new tiers
DELETE FROM subscription_tiers WHERE name IN ('starter', 'pro', 'custom');

-- Restore old tiers
INSERT INTO subscription_tiers (id, name, display_name, description, price_monthly, price_yearly, max_team_members, features, is_active, sort_order) VALUES
('tier-basic', 'basic', 'Basic', 'Perfect for small teams getting started', 0, 0, 5, '["document_management", "basic_workflows", "email_notifications"]'::jsonb, true, 1),
('tier-professional', 'professional', 'Professional', 'Advanced features for growing organizations', 50, 500, 25, '["document_management", "advanced_workflows", "email_notifications", "custom_roles", "analytics", "api_access"]'::jsonb, true, 2),
('tier-enterprise', 'enterprise', 'Enterprise', 'Full-featured solution for large organizations', 150, 1500, 100, '["document_management", "advanced_workflows", "email_notifications", "custom_roles", "analytics", "api_access", "sso", "audit_logs", "priority_support"]'::jsonb, true, 3),
('tier-unlimited', 'unlimited', 'Unlimited', 'No limits for enterprise customers', 500, 5000, -1, '["document_management", "advanced_workflows", "email_notifications", "custom_roles", "analytics", "api_access", "sso", "audit_logs", "priority_support", "custom_integrations", "dedicated_support"]'::jsonb, true, 4)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- STEP 3: Restore old subscription features
-- ============================================================================

-- Delete new features
DELETE FROM subscription_features;

-- Restore old features
INSERT INTO subscription_features (id, name, display_name, description, category, is_active) VALUES
('feat-doc-mgmt', 'document_management', 'Document Management', 'Create, edit, and manage documents', 'core', true),
('feat-basic-workflows', 'basic_workflows', 'Basic Workflows', 'Simple approval workflows', 'workflow', true),
('feat-advanced-workflows', 'advanced_workflows', 'Advanced Workflows', 'Complex multi-stage workflows with conditions', 'workflow', true),
('feat-email-notifications', 'email_notifications', 'Email Notifications', 'Automated email notifications', 'communication', true),
('feat-custom-roles', 'custom_roles', 'Custom Roles', 'Create and manage custom user roles', 'security', true),
('feat-analytics', 'analytics', 'Analytics & Reporting', 'Detailed analytics and custom reports', 'analytics', true),
('feat-api-access', 'api_access', 'API Access', 'REST API access for integrations', 'integration', true),
('feat-sso', 'sso', 'Single Sign-On', 'SAML/OAuth SSO integration', 'security', true),
('feat-audit-logs', 'audit_logs', 'Audit Logs', 'Comprehensive audit trail', 'security', true),
('feat-priority-support', 'priority_support', 'Priority Support', '24/7 priority customer support', 'support', true),
('feat-custom-integrations', 'custom_integrations', 'Custom Integrations', 'Custom API integrations and webhooks', 'integration', true),
('feat-dedicated-support', 'dedicated_support', 'Dedicated Support', 'Dedicated customer success manager', 'support', true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- STEP 4: Restore subscription_tiers table structure
-- ============================================================================

-- Remove new columns
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_workspaces;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_documents;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_workflows;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_custom_roles;

-- Restore old columns
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS storage_limit_gb INTEGER NOT NULL DEFAULT 1;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_organizations INTEGER;

-- Rename max_team_members back to max_users
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='subscription_tiers' AND column_name='max_team_members') THEN
        ALTER TABLE subscription_tiers RENAME COLUMN max_team_members TO max_users;
    END IF;
END $$;

-- Update restored tiers with old column values
UPDATE subscription_tiers SET storage_limit_gb = 1 WHERE name = 'basic';
UPDATE subscription_tiers SET storage_limit_gb = 10 WHERE name = 'professional';
UPDATE subscription_tiers SET storage_limit_gb = 50 WHERE name = 'enterprise';
UPDATE subscription_tiers SET storage_limit_gb = -1 WHERE name = 'unlimited';

-- Drop new index
DROP INDEX IF EXISTS idx_subscription_tiers_name;

-- ============================================================================
-- STEP 5: Restore organization_limit_overrides table structure
-- ============================================================================

-- Remove new columns
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_workspaces;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_documents;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_workflows;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_custom_roles;

-- Restore old columns
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS storage_limit_gb INTEGER;

-- Rename max_team_members back to max_users
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='organization_limit_overrides' AND column_name='max_team_members') THEN
        ALTER TABLE organization_limit_overrides RENAME COLUMN max_team_members TO max_users;
    END IF;
END $$;

-- Drop new indexes
DROP INDEX IF EXISTS idx_organization_overrides_unique;
DROP INDEX IF EXISTS idx_subscription_features_name;

-- ============================================================================
-- STEP 6: Create audit log entry for rollback
-- ============================================================================

INSERT INTO admin_audit_logs (id, action, details, reason, admin_user_id, created_at)
VALUES (
    'audit-rollback-014-' || EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::TEXT,
    'system_rollback',
    jsonb_build_object(
        'migration', '014_subscription_tier_system',
        'description', 'Rolled back to 4-tier system (basic, professional, enterprise, unlimited)',
        'rollback_date', CURRENT_TIMESTAMP
    ),
    'Rollback of migration 014_subscription_tier_system',
    'system',
    CURRENT_TIMESTAMP
);

RAISE NOTICE 'Migration 014 rolled back successfully.';
