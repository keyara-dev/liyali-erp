-- Migration 029: Create subscription_feature_requirements table and fix organization_has_feature()
--
-- Migration 010 dropped the subscription-focused feature_flags table and replaced it
-- with an admin UI flag management table that has a completely different schema.
-- The organization_has_feature() function (created in 007/027) was left querying
-- a table that no longer has plan_requirements / is_trial_allowed columns.
--
-- Fix: create a dedicated subscription_feature_requirements table and update the function.

-- ============================================================================
-- 1. Create the subscription feature requirements table
-- ============================================================================

CREATE TABLE IF NOT EXISTS subscription_feature_requirements (
    name               VARCHAR(100) PRIMARY KEY,
    description        TEXT,
    plan_requirements  JSONB NOT NULL DEFAULT '[]',
    is_trial_allowed   BOOLEAN NOT NULL DEFAULT FALSE,
    is_enterprise_only BOOLEAN NOT NULL DEFAULT FALSE,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 2. Seed feature requirements (all features used by FeatureGate in the frontend)
-- ============================================================================

INSERT INTO subscription_feature_requirements (name, description, plan_requirements, is_trial_allowed, is_enterprise_only) VALUES
-- Starter (free during trial)
('core_workflows',         'Basic approval workflows',                  '["STARTER_PLAN", "PRO_PLAN", "ENTERPRISE"]'::jsonb, true,  false),
('document_verification',  'Document authenticity verification',        '["STARTER_PLAN", "PRO_PLAN", "ENTERPRISE"]'::jsonb, true,  false),
('standard_analytics',     'Standard reporting',                        '["STARTER_PLAN", "PRO_PLAN", "ENTERPRISE"]'::jsonb, true,  false),
-- Pro tier
('custom_roles',           'Create and manage custom user roles',       '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('audit_logs_90_days',     'Audit trail with 90-day retention',         '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('advanced_workflows',     'Complex multi-stage conditional workflows', '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             true,  false),
('email_notifications',    'Automated email notifications',             '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             true,  false),
('advanced_analytics',     'Advanced reporting and analytics',          '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('data_export',            'Export data to CSV/Excel',                  '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('priority_support',       'Priority customer support',                 '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('multi_currency',         'Support for multiple currencies',           '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('advanced_reporting',     'Custom report builder',                     '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('workflow_templates',     'Pre-built workflow templates',              '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             true,  false),
('bulk_operations',        'Batch process documents',                   '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('offline_capabilities',   'Work offline and sync when connected',      '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
('api_access',             'Access to REST API endpoints',              '["PRO_PLAN", "ENTERPRISE"]'::jsonb,             false, false),
-- Enterprise only
('webhooks',               'Real-time event webhooks',                  '["ENTERPRISE"]'::jsonb,                         false, false),
('custom_fields',          'Add custom fields to documents',            '["ENTERPRISE"]'::jsonb,                         false, false),
('dedicated_instance',     'Dedicated server instance',                 '["ENTERPRISE"]'::jsonb,                         false, true),
('sla_guarantees',         'Service Level Agreement guarantees',        '["ENTERPRISE"]'::jsonb,                         false, true),
('custom_integrations',    'Custom third-party integrations',           '["ENTERPRISE"]'::jsonb,                         false, true),
('models_modification',    'Create and modify data models',             '["ENTERPRISE"]'::jsonb,                         false, true),
('unlimited_users',        'No user limit restrictions',                '["ENTERPRISE"]'::jsonb,                         false, true),
('audit_logs_unlimited',   'Unlimited audit log retention',             '["ENTERPRISE"]'::jsonb,                         false, true)
ON CONFLICT (name) DO UPDATE SET
    description        = EXCLUDED.description,
    plan_requirements  = EXCLUDED.plan_requirements,
    is_trial_allowed   = EXCLUDED.is_trial_allowed,
    is_enterprise_only = EXCLUDED.is_enterprise_only,
    updated_at         = CURRENT_TIMESTAMP;

-- ============================================================================
-- 3. Rewrite organization_has_feature() to use the new table
-- ============================================================================

CREATE OR REPLACE FUNCTION organization_has_feature(org_id VARCHAR(255), feature_name VARCHAR(100))
RETURNS BOOLEAN AS $$
DECLARE
    org_status     VARCHAR(50);
    org_tier       VARCHAR(50);
    org_plan_slug  VARCHAR(50);
    org_trial_end  TIMESTAMP;
    org_grace_end  TIMESTAMP;
    feature_plans  JSONB;
    is_trial_ok    BOOLEAN;
    current_time   TIMESTAMP := CURRENT_TIMESTAMP;
BEGIN
    -- Get organization subscription details
    SELECT
        o.subscription_status,
        COALESCE(o.subscription_tier, o.tier, 'starter'),
        sp.slug,
        o.trial_end_date,
        o.grace_period_ends_at
    INTO org_status, org_tier, org_plan_slug, org_trial_end, org_grace_end
    FROM organizations o
    LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
    WHERE o.id = org_id;

    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;

    -- PRO/custom tier: bypass trial checks entirely
    IF org_tier IN ('pro', 'custom') THEN
        IF org_plan_slug IS NULL THEN
            org_plan_slug := CASE org_tier
                WHEN 'pro'    THEN 'PRO_PLAN'
                WHEN 'custom' THEN 'ENTERPRISE'
                ELSE org_tier
            END;
        END IF;
        SELECT plan_requirements
        INTO feature_plans
        FROM subscription_feature_requirements
        WHERE name = feature_name;
        IF feature_plans IS NULL THEN
            RETURN FALSE;
        END IF;
        RETURN feature_plans ? org_plan_slug
            OR feature_plans ? 'PRO_PLAN'
            OR feature_plans ? 'ENTERPRISE';
    END IF;

    -- Get feature requirements for non-paid tiers
    SELECT plan_requirements, is_trial_allowed
    INTO feature_plans, is_trial_ok
    FROM subscription_feature_requirements
    WHERE name = feature_name;

    IF feature_plans IS NULL THEN
        RETURN FALSE;
    END IF;

    -- Trial logic
    IF org_status = 'trial' THEN
        IF org_trial_end IS NOT NULL AND current_time > org_trial_end THEN
            -- Grace period: basic features only
            IF org_grace_end IS NOT NULL AND current_time <= org_grace_end THEN
                RETURN feature_name IN ('core_workflows', 'document_verification', 'standard_analytics');
            ELSE
                RETURN FALSE;
            END IF;
        ELSE
            -- Active trial
            IF is_trial_ok THEN
                RETURN TRUE;
            ELSE
                RETURN feature_plans ? COALESCE(org_plan_slug, 'STARTER_PLAN');
            END IF;
        END IF;
    END IF;

    -- Active subscription
    IF org_status = 'active' THEN
        RETURN feature_plans ? COALESCE(org_plan_slug, 'STARTER_PLAN');
    END IF;

    -- Past due: grace period basic access
    IF org_status = 'past_due' THEN
        IF org_grace_end IS NOT NULL AND current_time <= org_grace_end THEN
            RETURN feature_name IN ('core_workflows', 'document_verification', 'standard_analytics');
        ELSE
            RETURN FALSE;
        END IF;
    END IF;

    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;
