-- Migration: 014_subscription_tier_system
-- Description: Update subscription system to 3-tier model (STARTER, PRO, CUSTOM) with database-driven configuration
-- Date: 2026-02-24
-- Purpose: Replace hardcoded tier logic with fully database-driven subscription management

-- ============================================================================
-- STEP 1: Update subscription_tiers table structure
-- ============================================================================

-- Remove storage_limit_gb column (no tracking mechanism exists)
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS storage_limit_gb;

-- Remove max_organizations column (not needed - workspaces are the limit)
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_organizations;

-- Rename max_users to max_team_members for clarity
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='subscription_tiers' AND column_name='max_users') THEN
        ALTER TABLE subscription_tiers RENAME COLUMN max_users TO max_team_members;
    END IF;
END $$;

-- Add new limit columns
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_workspaces INTEGER NOT NULL DEFAULT 1;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_documents INTEGER NOT NULL DEFAULT 100;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_workflows INTEGER NOT NULL DEFAULT 1;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_custom_roles INTEGER NOT NULL DEFAULT 0;

-- Add index for tier name lookups
CREATE INDEX IF NOT EXISTS idx_subscription_tiers_name ON subscription_tiers(name);

-- ============================================================================
-- STEP 2: Update organization_limit_overrides table structure
-- ============================================================================

-- Remove storage_limit_gb column
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS storage_limit_gb;

-- Rename max_users to max_team_members
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='organization_limit_overrides' AND column_name='max_users') THEN
        ALTER TABLE organization_limit_overrides RENAME COLUMN max_users TO max_team_members;
    END IF;
END $$;

-- Add new limit override columns
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_workspaces INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_documents INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_workflows INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_custom_roles INTEGER;

-- Add unique constraint to ensure one override per organization
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_overrides_unique ON organization_limit_overrides(organization_id);

-- Add index for feature name lookups
CREATE INDEX IF NOT EXISTS idx_subscription_features_name ON subscription_features(name);

-- ============================================================================
-- STEP 3: Migrate existing tier data to new 3-tier system
-- ============================================================================

-- Backup existing tier data in audit log
INSERT INTO admin_audit_logs (id, action, old_value, new_value, details, reason, admin_user_id, created_at)
SELECT 
    'audit-tier-migration-' || id,
    'tier_migration',
    name,
    CASE 
        WHEN name = 'basic' THEN 'starter'
        WHEN name = 'professional' THEN 'pro'
        WHEN name IN ('enterprise', 'unlimited') THEN 'custom'
        ELSE name
    END,
    jsonb_build_object(
        'old_tier_id', id,
        'old_display_name', display_name,
        'old_features', features,
        'migration_date', CURRENT_TIMESTAMP
    ),
    'Automated migration to 3-tier system (STARTER, PRO, CUSTOM)',
    'system',
    CURRENT_TIMESTAMP
FROM subscription_tiers
WHERE name IN ('basic', 'professional', 'enterprise', 'unlimited');

-- Delete old tiers (will be replaced with new ones)
DELETE FROM subscription_tiers WHERE name IN ('basic', 'professional', 'enterprise', 'unlimited');

-- ============================================================================
-- STEP 4: Insert new 3-tier system data
-- ============================================================================

-- Insert STARTER tier
INSERT INTO subscription_tiers (
    id, name, display_name, description, 
    price_monthly, price_yearly,
    max_workspaces, max_team_members, max_documents, max_workflows, max_custom_roles,
    features, is_active, sort_order
) VALUES (
    'tier-starter',
    'starter',
    'Starter',
    'Perfect for small teams getting started',
    0,
    0,
    1,
    10,
    200,
    3,
    0,
    '["document_management", "basic_workflows", "in_app_notifications", "standard_reports", "user_management", "department_management", "vendor_management", "budget_tracking", "mobile_web_access", "email_support"]'::jsonb,
    true,
    1
) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    price_monthly = EXCLUDED.price_monthly,
    price_yearly = EXCLUDED.price_yearly,
    max_workspaces = EXCLUDED.max_workspaces,
    max_team_members = EXCLUDED.max_team_members,
    max_documents = EXCLUDED.max_documents,
    max_workflows = EXCLUDED.max_workflows,
    max_custom_roles = EXCLUDED.max_custom_roles,
    features = EXCLUDED.features,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

-- Insert PRO tier
INSERT INTO subscription_tiers (
    id, name, display_name, description,
    price_monthly, price_yearly,
    max_workspaces, max_team_members, max_documents, max_workflows, max_custom_roles,
    features, is_active, sort_order
) VALUES (
    'tier-pro',
    'pro',
    'Pro',
    'Advanced features for growing organizations',
    99,
    990,
    5,
    50,
    500,
    20,
    10,
    '["document_management", "basic_workflows", "in_app_notifications", "standard_reports", "user_management", "department_management", "vendor_management", "budget_tracking", "mobile_web_access", "email_support", "advanced_workflows", "email_notifications", "custom_roles", "advanced_analytics", "data_export", "priority_support", "audit_logs_90_days", "multi_currency", "advanced_reporting", "workflow_templates"]'::jsonb,
    true,
    2
) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    price_monthly = EXCLUDED.price_monthly,
    price_yearly = EXCLUDED.price_yearly,
    max_workspaces = EXCLUDED.max_workspaces,
    max_team_members = EXCLUDED.max_team_members,
    max_documents = EXCLUDED.max_documents,
    max_workflows = EXCLUDED.max_workflows,
    max_custom_roles = EXCLUDED.max_custom_roles,
    features = EXCLUDED.features,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

-- Insert CUSTOM tier
INSERT INTO subscription_tiers (
    id, name, display_name, description,
    price_monthly, price_yearly,
    max_workspaces, max_team_members, max_documents, max_workflows, max_custom_roles,
    features, is_active, sort_order
) VALUES (
    'tier-custom',
    'custom',
    'Custom',
    'Enterprise solution with configurable limits',
    499,
    4990,
    -1,
    -1,
    -1,
    -1,
    -1,
    '["document_management", "basic_workflows", "in_app_notifications", "standard_reports", "user_management", "department_management", "vendor_management", "budget_tracking", "mobile_web_access", "email_support", "advanced_workflows", "email_notifications", "custom_roles", "advanced_analytics", "data_export", "priority_support", "audit_logs_90_days", "multi_currency", "advanced_reporting", "workflow_templates", "webhooks", "custom_fields", "bulk_operations", "api_access", "dedicated_support_manager", "sla_guarantees", "audit_logs_unlimited", "custom_development", "professional_services", "dedicated_support", "custom_training"]'::jsonb,
    true,
    3
) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    price_monthly = EXCLUDED.price_monthly,
    price_yearly = EXCLUDED.price_yearly,
    max_workspaces = EXCLUDED.max_workspaces,
    max_team_members = EXCLUDED.max_team_members,
    max_documents = EXCLUDED.max_documents,
    max_workflows = EXCLUDED.max_workflows,
    max_custom_roles = EXCLUDED.max_custom_roles,
    features = EXCLUDED.features,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

-- ============================================================================
-- STEP 5: Update subscription_features with complete feature set
-- ============================================================================

-- Delete old features (will be replaced with comprehensive set)
DELETE FROM subscription_features;

-- Insert Core Features (STARTER tier)
INSERT INTO subscription_features (id, name, display_name, description, category, is_active) VALUES
('feat-doc-mgmt', 'document_management', 'Document Management', 'Create, edit, and manage documents', 'core', true),
('feat-basic-workflows', 'basic_workflows', 'Basic Workflows', 'Simple linear approval workflows (1-3 stages)', 'workflow', true),
('feat-in-app-notif', 'in_app_notifications', 'In-App Notifications', 'Real-time in-app notifications', 'core', true),
('feat-std-reports', 'standard_reports', 'Standard Reports', 'Pre-built standard reports', 'analytics', true),
('feat-user-mgmt', 'user_management', 'User Management', 'Manage users with system roles', 'core', true),
('feat-dept-mgmt', 'department_management', 'Department Management', 'Organize users into departments', 'core', true),
('feat-vendor-mgmt', 'vendor_management', 'Vendor Management', 'Maintain vendor database', 'core', true),
('feat-budget-track', 'budget_tracking', 'Budget Tracking', 'Track budget utilization', 'core', true),
('feat-mobile-web', 'mobile_web_access', 'Mobile Web Access', 'Responsive mobile web interface', 'core', true),
('feat-email-support', 'email_support', 'Email Support', 'Standard email support', 'support', true)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    is_active = EXCLUDED.is_active;

-- Insert PRO Features
INSERT INTO subscription_features (id, name, display_name, description, category, is_active) VALUES
('feat-adv-workflows', 'advanced_workflows', 'Advanced Workflows', 'Complex multi-stage workflows with conditional routing', 'workflow', true),
('feat-email-notif', 'email_notifications', 'Email Notifications', 'Automated email notifications', 'core', true),
('feat-custom-roles', 'custom_roles', 'Custom Roles & Permissions', 'Create organization-specific roles', 'security', true),
('feat-adv-analytics', 'advanced_analytics', 'Advanced Analytics & Dashboards', 'Custom dashboards and analytics', 'analytics', true),
('feat-data-export', 'data_export', 'Data Export', 'Export data to CSV/Excel', 'analytics', true),
('feat-priority-support', 'priority_support', 'Priority Support', '24/7 priority customer support', 'support', true),
('feat-audit-90', 'audit_logs_90_days', 'Audit Logs (90 days)', 'Audit trail with 90-day retention', 'security', true),
('feat-multi-currency', 'multi_currency', 'Multi-Currency Support', 'Support for multiple currencies', 'core', true),
('feat-adv-reporting', 'advanced_reporting', 'Advanced Reporting', 'Custom report builder', 'analytics', true),
('feat-workflow-templates', 'workflow_templates', 'Workflow Templates', 'Pre-built workflow templates', 'workflow', true)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    is_active = EXCLUDED.is_active;

-- Insert CUSTOM Features
INSERT INTO subscription_features (id, name, display_name, description, category, is_active) VALUES
('feat-webhooks', 'webhooks', 'Webhooks', 'Real-time event webhooks', 'integration', true),
('feat-custom-fields', 'custom_fields', 'Custom Fields', 'Add custom fields to documents', 'customization', true),
('feat-bulk-ops', 'bulk_operations', 'Bulk Operations', 'Batch process documents', 'core', true),
('feat-api-access', 'api_access', 'API Access', 'Full REST API access', 'integration', true),
('feat-dedicated-mgr', 'dedicated_support_manager', 'Dedicated Support Manager', 'Dedicated customer success manager', 'support', true),
('feat-sla', 'sla_guarantees', 'SLA Guarantees', 'Service level agreement guarantees', 'support', true),
('feat-audit-unlimited', 'audit_logs_unlimited', 'Audit Logs (Unlimited)', 'Unlimited audit log retention', 'security', true),
('feat-custom-dev', 'custom_development', 'Custom Development', 'Bespoke feature development', 'customization', true),
('feat-prof-services', 'professional_services', 'Professional Services', 'Implementation consulting', 'support', true),
('feat-dedicated-support', 'dedicated_support', 'Dedicated Support', '24/7 dedicated support team', 'support', true),
('feat-custom-training', 'custom_training', 'Custom Training', 'Tailored training sessions', 'support', true)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    is_active = EXCLUDED.is_active;

-- ============================================================================
-- STEP 6: Migrate existing organizations to new tier names
-- ============================================================================

-- Map old tier names to new tier names
UPDATE organizations
SET subscription_tier = CASE subscription_tier
    WHEN 'basic' THEN 'starter'
    WHEN 'professional' THEN 'pro'
    WHEN 'enterprise' THEN 'custom'
    WHEN 'unlimited' THEN 'custom'
    ELSE 'starter' -- Default to starter for any unknown tiers
END
WHERE subscription_tier IN ('basic', 'professional', 'enterprise', 'unlimited');

-- Set default tier for organizations without a tier
UPDATE organizations
SET subscription_tier = 'starter'
WHERE subscription_tier IS NULL OR subscription_tier = '';

-- Ensure all organizations have trial dates if in trial status
UPDATE organizations
SET 
    trial_start_date = COALESCE(trial_start_date, created_at),
    trial_end_date = COALESCE(trial_end_date, created_at + INTERVAL '30 days')
WHERE subscription_status = 'trial' AND (trial_start_date IS NULL OR trial_end_date IS NULL);

-- ============================================================================
-- STEP 7: Create audit log entry for migration
-- ============================================================================

INSERT INTO admin_audit_logs (id, action, details, reason, admin_user_id, created_at)
VALUES (
    'audit-migration-014-' || EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::TEXT,
    'system_migration',
    jsonb_build_object(
        'migration', '014_subscription_tier_system',
        'description', 'Migrated to 3-tier system (STARTER, PRO, CUSTOM)',
        'changes', jsonb_build_array(
            'Removed storage_limit_gb column',
            'Renamed max_users to max_team_members',
            'Added max_workspaces, max_documents, max_workflows, max_custom_roles',
            'Replaced 4 tiers with 3 tiers',
            'Updated all features to match new tier structure',
            'Migrated existing organizations to new tier names'
        ),
        'organizations_migrated', (SELECT COUNT(*) FROM organizations WHERE subscription_tier IN ('starter', 'pro', 'custom')),
        'completed_at', CURRENT_TIMESTAMP
    ),
    'Automated database migration to implement database-driven 3-tier subscription system',
    'system',
    CURRENT_TIMESTAMP
);

-- ============================================================================
-- STEP 8: Verify migration success
-- ============================================================================

-- This will raise an error if any organization has an invalid tier
DO $$
DECLARE
    invalid_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO invalid_count
    FROM organizations o
    WHERE o.subscription_tier NOT IN (SELECT name FROM subscription_tiers WHERE is_active = true);
    
    IF invalid_count > 0 THEN
        RAISE EXCEPTION 'Migration validation failed: % organizations have invalid subscription tiers', invalid_count;
    END IF;
    
    RAISE NOTICE 'Migration 014 completed successfully. All organizations have valid subscription tiers.';
END $$;
