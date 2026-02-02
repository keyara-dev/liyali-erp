-- ============================================================================
-- LIYALI GATEWAY - SUBSCRIPTION SYSTEM ROLLBACK
-- Migration: 007_subscription_system.down
-- Description: Rollback subscription system changes
-- Version: 1.0.0
-- Date: February 1, 2026
-- ============================================================================

-- Drop triggers
DROP TRIGGER IF EXISTS auto_start_trial ON organizations;

-- Drop functions
DROP FUNCTION IF EXISTS trigger_start_organization_trial();
DROP FUNCTION IF EXISTS extend_organization_trial(VARCHAR, INTEGER, VARCHAR);
DROP FUNCTION IF EXISTS organization_has_feature(VARCHAR, VARCHAR);
DROP FUNCTION IF EXISTS start_organization_trial(VARCHAR);

-- Drop views
DROP VIEW IF EXISTS organization_subscription_details;

-- Drop indexes
DROP INDEX IF EXISTS idx_subscription_audit_action;
DROP INDEX IF EXISTS idx_subscription_audit_performed_at;
DROP INDEX IF EXISTS idx_subscription_audit_organization_id;
DROP INDEX IF EXISTS idx_feature_flags_active;
DROP INDEX IF EXISTS idx_feature_flags_name;
DROP INDEX IF EXISTS idx_org_subscriptions_stripe_id;
DROP INDEX IF EXISTS idx_org_subscriptions_status;
DROP INDEX IF EXISTS idx_org_subscriptions_plan_id;
DROP INDEX IF EXISTS idx_org_subscriptions_organization_id;
DROP INDEX IF EXISTS idx_organizations_current_plan_id;
DROP INDEX IF EXISTS idx_organizations_grace_period_ends_at;
DROP INDEX IF EXISTS idx_organizations_trial_end_date;
DROP INDEX IF EXISTS idx_organizations_subscription_status;

-- Drop tables
DROP TABLE IF EXISTS subscription_audit_logs;
DROP TABLE IF EXISTS organization_subscriptions;
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS subscription_plans;

-- Remove columns from organizations table
ALTER TABLE organizations 
DROP CONSTRAINT IF EXISTS check_organization_subscription_status,
DROP CONSTRAINT IF EXISTS fk_organizations_current_plan,
DROP COLUMN IF EXISTS subscription_metadata,
DROP COLUMN IF EXISTS max_users_allowed,
DROP COLUMN IF EXISTS grace_period_ends_at,
DROP COLUMN IF EXISTS billing_cycle_end,
DROP COLUMN IF EXISTS billing_cycle_start,
DROP COLUMN IF EXISTS subscription_status,
DROP COLUMN IF EXISTS current_plan_id,
DROP COLUMN IF EXISTS trial_end_date,
DROP COLUMN IF EXISTS trial_start_date;