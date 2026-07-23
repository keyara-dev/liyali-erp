-- ============================================================================
-- ROLLBACK: 002_subscription_system
-- ============================================================================

ALTER TABLE organizations DROP CONSTRAINT IF EXISTS fk_organizations_current_plan;

DROP VIEW IF EXISTS organization_subscription_details CASCADE;

DROP FUNCTION IF EXISTS organization_has_feature(VARCHAR, VARCHAR) CASCADE;
DROP FUNCTION IF EXISTS start_organization_trial(VARCHAR) CASCADE;
DROP FUNCTION IF EXISTS trigger_start_organization_trial() CASCADE;
DROP FUNCTION IF EXISTS track_subscription_tier_changes() CASCADE;
DROP FUNCTION IF EXISTS track_trial_conversions() CASCADE;

DROP TABLE IF EXISTS subscription_feature_requirements CASCADE;
DROP TABLE IF EXISTS subscription_events CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS admin_audit_logs CASCADE;
DROP TABLE IF EXISTS organization_limit_overrides CASCADE;
DROP TABLE IF EXISTS subscription_features CASCADE;
DROP TABLE IF EXISTS subscription_tiers CASCADE;
DROP TABLE IF EXISTS subscription_audit_logs CASCADE;
DROP TABLE IF EXISTS organization_subscriptions CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;
