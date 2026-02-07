-- Migration: 012_subscription_management_system
-- Description: Create comprehensive subscription management system for admin console
-- Date: 2024-02-05

-- Create subscription tiers table
CREATE TABLE IF NOT EXISTS subscription_tiers (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL,
    price_monthly REAL NOT NULL DEFAULT 0,
    price_yearly REAL NOT NULL DEFAULT 0,
    max_users INTEGER NOT NULL DEFAULT 1,
    max_organizations INTEGER,
    storage_limit_gb INTEGER NOT NULL DEFAULT 1,
    features TEXT NOT NULL DEFAULT '[]', -- JSON array of feature names
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create subscription features table
CREATE TABLE IF NOT EXISTS subscription_features (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create organization limit overrides table
CREATE TABLE IF NOT EXISTS organization_limit_overrides (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    max_users INTEGER,
    storage_limit_gb INTEGER,
    features TEXT, -- JSON array of additional features
    reason TEXT NOT NULL,
    admin_user_id TEXT NOT NULL,
    expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Create admin audit logs table for subscription changes
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id TEXT PRIMARY KEY,
    organization_id TEXT,
    action TEXT NOT NULL, -- 'tier_change', 'limit_override', 'trial_reset', etc.
    old_value TEXT,
    new_value TEXT,
    details TEXT, -- JSON object with additional details
    reason TEXT,
    admin_user_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Add subscription-related columns to organizations table if they don't exist
ALTER TABLE organizations ADD COLUMN subscription_tier TEXT DEFAULT 'basic';
ALTER TABLE organizations ADD COLUMN subscription_status TEXT DEFAULT 'trial';
ALTER TABLE organizations ADD COLUMN trial_start_date DATETIME;
ALTER TABLE organizations ADD COLUMN trial_end_date DATETIME;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_subscription_tiers_active ON subscription_tiers(is_active);
CREATE INDEX IF NOT EXISTS idx_subscription_tiers_sort ON subscription_tiers(sort_order);
CREATE INDEX IF NOT EXISTS idx_subscription_features_category ON subscription_features(category);
CREATE INDEX IF NOT EXISTS idx_subscription_features_active ON subscription_features(is_active);
CREATE INDEX IF NOT EXISTS idx_organization_overrides_org ON organization_limit_overrides(organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_overrides_expires ON organization_limit_overrides(expires_at);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_org ON admin_audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_action ON admin_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_created ON admin_audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_organizations_subscription_tier ON organizations(subscription_tier);
CREATE INDEX IF NOT EXISTS idx_organizations_subscription_status ON organizations(subscription_status);
CREATE INDEX IF NOT EXISTS idx_organizations_trial_end ON organizations(trial_end_date);

-- Insert default subscription tiers
INSERT OR IGNORE INTO subscription_tiers (id, name, display_name, description, price_monthly, price_yearly, max_users, storage_limit_gb, features, is_active, sort_order) VALUES
('tier-basic', 'basic', 'Basic', 'Perfect for small teams getting started', 0, 0, 5, 1, '["document_management", "basic_workflows", "email_notifications"]', true, 1),
('tier-professional', 'professional', 'Professional', 'Advanced features for growing organizations', 50, 500, 25, 10, '["document_management", "advanced_workflows", "email_notifications", "custom_roles", "analytics", "api_access"]', true, 2),
('tier-enterprise', 'enterprise', 'Enterprise', 'Full-featured solution for large organizations', 150, 1500, 100, 50, '["document_management", "advanced_workflows", "email_notifications", "custom_roles", "analytics", "api_access", "sso", "audit_logs", "priority_support"]', true, 3),
('tier-unlimited', 'unlimited', 'Unlimited', 'No limits for enterprise customers', 500, 5000, -1, -1, '["document_management", "advanced_workflows", "email_notifications", "custom_roles", "analytics", "api_access", "sso", "audit_logs", "priority_support", "custom_integrations", "dedicated_support"]', true, 4);

-- Insert default subscription features
INSERT OR IGNORE INTO subscription_features (id, name, display_name, description, category, is_active) VALUES
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
('feat-dedicated-support', 'dedicated_support', 'Dedicated Support', 'Dedicated customer success manager', 'support', true);

-- Update existing organizations to have proper trial dates if they don't exist
UPDATE organizations 
SET 
    trial_start_date = created_at,
    trial_end_date = datetime(created_at, '+30 days')
WHERE trial_start_date IS NULL AND subscription_status = 'trial';

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_subscription_tiers_updated_at
    AFTER UPDATE ON subscription_tiers
    FOR EACH ROW
BEGIN
    UPDATE subscription_tiers SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_organization_overrides_updated_at
    AFTER UPDATE ON organization_limit_overrides
    FOR EACH ROW
BEGIN
    UPDATE organization_limit_overrides SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;