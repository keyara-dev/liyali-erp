-- Migration 027: PRO/custom tier overrides trial — fix feature gate + existing data
-- ============================================================================
-- 1. Fix existing orgs already on pro/custom that still have trial status
-- ============================================================================

UPDATE organizations
SET subscription_status    = 'active',
    trial_end_date         = NULL,
    grace_period_ends_at   = NULL,
    updated_at             = CURRENT_TIMESTAMP
WHERE subscription_tier IN ('pro', 'custom')
  AND (subscription_status = 'trial'
    OR subscription_status IS NULL
    OR trial_end_date IS NOT NULL
    OR grace_period_ends_at IS NOT NULL);

-- ============================================================================
-- 2. Update organization_has_feature() to bypass trial logic for paid tiers
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
    is_trial_allowed BOOLEAN;
    current_time   TIMESTAMP := CURRENT_TIMESTAMP;
BEGIN
    -- Get organization subscription details
    SELECT
        o.subscription_status,
        o.subscription_tier,
        sp.slug,
        o.trial_end_date,
        o.grace_period_ends_at
    INTO org_status, org_tier, org_plan_slug, org_trial_end, org_grace_end
    FROM organizations o
    LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
    WHERE o.id = org_id;

    -- If organization not found, deny access
    IF org_status IS NULL AND org_tier IS NULL THEN
        RETURN FALSE;
    END IF;

    -- PRO/custom tier: bypass trial checks entirely, treat as fully active
    IF org_tier IN ('pro', 'custom') THEN
        -- Derive plan slug from tier so plan-level feature checks still apply
        IF org_plan_slug IS NULL THEN
            org_plan_slug := CASE org_tier
                WHEN 'pro'    THEN 'PRO_PLAN'
                WHEN 'custom' THEN 'ENTERPRISE'
                ELSE org_tier
            END;
        END IF;
        SELECT plan_requirements, is_trial_allowed
        INTO feature_plans, is_trial_allowed
        FROM feature_flags
        WHERE name = feature_name AND is_active = true;
        IF feature_plans IS NULL THEN
            RETURN FALSE;
        END IF;
        RETURN feature_plans ? org_plan_slug OR feature_plans ? 'PRO_PLAN' OR feature_plans ? 'ENTERPRISE';
    END IF;

    -- Get feature requirements
    SELECT plan_requirements, is_trial_allowed
    INTO feature_plans, is_trial_allowed
    FROM feature_flags
    WHERE name = feature_name AND is_active = true;

    -- If feature not found, deny access
    IF feature_plans IS NULL THEN
        RETURN FALSE;
    END IF;

    -- Check if organization is in trial
    IF org_status = 'trial' THEN
        -- If trial expired, check grace period
        IF current_time > org_trial_end THEN
            -- If in grace period, allow read-only access (basic features only)
            IF org_grace_end IS NOT NULL AND current_time <= org_grace_end THEN
                RETURN feature_name IN ('core_workflows', 'document_verification', 'standard_analytics');
            ELSE
                -- Trial expired, no grace period
                RETURN FALSE;
            END IF;
        ELSE
            -- Trial active, check if feature is allowed for trial
            IF is_trial_allowed THEN
                RETURN TRUE;
            ELSE
                RETURN feature_plans ? org_plan_slug;
            END IF;
        END IF;
    END IF;

    -- For active subscriptions, check plan requirements
    IF org_status = 'active' THEN
        RETURN feature_plans ? org_plan_slug;
    END IF;

    -- For past_due status, allow basic access during grace period
    IF org_status = 'past_due' THEN
        IF org_grace_end IS NOT NULL AND current_time <= org_grace_end THEN
            RETURN feature_name IN ('core_workflows', 'document_verification', 'standard_analytics');
        ELSE
            RETURN FALSE;
        END IF;
    END IF;

    -- For canceled or expired, deny access
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;
