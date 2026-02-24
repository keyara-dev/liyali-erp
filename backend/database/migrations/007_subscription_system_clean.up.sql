-- ============================================================================
-- LIYALI GATEWAY - SUBSCRIPTION MANAGEMENT SYSTEM (CLEAN VERSION)
-- Migration: 008_subscription_system_clean
-- Description: Complete subscription system with trials, plans, and feature gating
-- Version: 1.0.0
-- Date: February 1, 2026
-- ============================================================================

-- ============================================================================
-- SUBSCRIPTION PLANS TABLE (Admin-managed)
-- ============================================================================
CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    price_monthly DECIMAL(10,2) DEFAULT 0.00,
    price_yearly DECIMAL(10,2) DEFAULT 0.00,
    features JSONB NOT NULL DEFAULT '[]',
    max_users INTEGER DEFAULT 50,
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT check_subscription_plan_slug CHECK (slug IN ('STARTER_PLAN', 'PRO_PLAN', 'ENTERPRISE'))
);

-- ============================================================================
-- FEATURE FLAGS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    plan_requirements JSONB NOT NULL DEFAULT '[]',
    is_trial_allowed BOOLEAN DEFAULT false,
    is_enterprise_only BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- ORGANIZATION SUBSCRIPTIONS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS organization_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    plan_id UUID NOT NULL,
    stripe_subscription_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'trial',
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT false,
    payment_failed_count INTEGER DEFAULT 0,
    last_payment_failed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_org_subscriptions_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_org_subscriptions_plan FOREIGN KEY (plan_id) REFERENCES subscription_plans(id),
    CONSTRAINT check_subscription_status CHECK (status IN ('trial', 'active', 'past_due', 'canceled', 'expired')),
    CONSTRAINT uk_org_subscription UNIQUE (organization_id)
);

-- ============================================================================
-- UPDATE ORGANIZATIONS TABLE FOR SUBSCRIPTION SYSTEM
-- ============================================================================

-- Add subscription-related columns to organizations table
ALTER TABLE organizations 
ADD COLUMN IF NOT EXISTS trial_start_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS trial_end_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS current_plan_id UUID,
ADD COLUMN IF NOT EXISTS subscription_status VARCHAR(50) DEFAULT 'trial',
ADD COLUMN IF NOT EXISTS billing_cycle_start TIMESTAMP,
ADD COLUMN IF NOT EXISTS billing_cycle_end TIMESTAMP,
ADD COLUMN IF NOT EXISTS grace_period_ends_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS max_users_allowed INTEGER DEFAULT 50,
ADD COLUMN IF NOT EXISTS subscription_metadata JSONB DEFAULT '{}';

-- Add foreign key constraint for current_plan_id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_organizations_current_plan'
    ) THEN
        ALTER TABLE organizations 
        ADD CONSTRAINT fk_organizations_current_plan 
        FOREIGN KEY (current_plan_id) REFERENCES subscription_plans(id);
    END IF;
END $$;

-- Add check constraint for subscription_status
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'check_organization_subscription_status'
    ) THEN
        ALTER TABLE organizations 
        ADD CONSTRAINT check_organization_subscription_status 
        CHECK (subscription_status IN ('trial', 'active', 'past_due', 'canceled', 'expired'));
    END IF;
END $$;

-- ============================================================================
-- SUBSCRIPTION AUDIT LOG TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS subscription_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    old_plan_id UUID,
    new_plan_id UUID,
    old_status VARCHAR(50),
    new_status VARCHAR(50),
    metadata JSONB DEFAULT '{}',
    performed_by VARCHAR(255),
    performed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_audit_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_audit_old_plan FOREIGN KEY (old_plan_id) REFERENCES subscription_plans(id),
    CONSTRAINT fk_audit_new_plan FOREIGN KEY (new_plan_id) REFERENCES subscription_plans(id)
);

-- ============================================================================
-- SEED DEFAULT SUBSCRIPTION PLANS
-- ============================================================================
INSERT INTO subscription_plans (name, slug, description, price_monthly, price_yearly, features, max_users, sort_order, metadata) VALUES
(
    'Starter Plan',
    'STARTER_PLAN',
    'Perfect for small teams getting started with procurement workflows',
    0.00,
    0.00,
    '[
        "Core procurement workflows",
        "Up to 50 users",
        "Single Workspace",
        "Document Verification (QR Codes and Doc Numbers)",
        "Standard analytics",
        "Notifications (Email & In-App)"
    ]'::jsonb,
    50,
    1,
    '{
        "offline_capabilities": false,
        "api_access": false,
        "custom_roles": false,
        "priority_support": false,
        "dedicated_instance": false,
        "sla_guarantees": false
    }'::jsonb
),
(
    'Pro Plan',
    'PRO_PLAN',
    'Advanced features for growing organizations',
    99.00,
    990.00,
    '[
        "Everything in Starter Plan",
        "Up to 200 users",
        "Custom Role management",
        "Offline capabilities",
        "Priority support",
        "Advanced analytics",
        "API Access"
    ]'::jsonb,
    200,
    2,
    '{
        "offline_capabilities": true,
        "api_access": true,
        "custom_roles": true,
        "priority_support": true,
        "dedicated_instance": false,
        "sla_guarantees": false
    }'::jsonb
),
(
    'Enterprise',
    'ENTERPRISE',
    'Complete solution for large organizations',
    0.00,
    0.00,
    '[
        "Everything in Pro Plan",
        "Unlimited users",
        "Dedicated instance",
        "Custom integrations",
        "SLA guarantees",
        "Dedicated success manager",
        "Models Creation/Modifications"
    ]'::jsonb,
    -1,
    3,
    '{
        "offline_capabilities": true,
        "api_access": true,
        "custom_roles": true,
        "priority_support": true,
        "dedicated_instance": true,
        "sla_guarantees": true,
        "custom_pricing": true
    }'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price_monthly = EXCLUDED.price_monthly,
    price_yearly = EXCLUDED.price_yearly,
    features = EXCLUDED.features,
    max_users = EXCLUDED.max_users,
    sort_order = EXCLUDED.sort_order,
    metadata = EXCLUDED.metadata,
    updated_at = CURRENT_TIMESTAMP;

-- ============================================================================
-- SEED FEATURE FLAGS
-- ============================================================================
INSERT INTO feature_flags (name, description, plan_requirements, is_trial_allowed, is_enterprise_only) VALUES
('custom_roles', 'Create and manage custom user roles', '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('offline_capabilities', 'Work offline and sync when connected', '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('api_access', 'Access to REST API endpoints', '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('priority_support', 'Priority customer support', '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('dedicated_instance', 'Dedicated server instance', '["ENTERPRISE"]'::jsonb, false, true),
('sla_guarantees', 'Service Level Agreement guarantees', '["ENTERPRISE"]'::jsonb, false, true),
('custom_integrations', 'Custom third-party integrations', '["ENTERPRISE"]'::jsonb, false, true),
('models_modification', 'Create and modify data models', '["ENTERPRISE"]'::jsonb, false, true),
('advanced_analytics', 'Advanced reporting and analytics', '["PRO_PLAN", "ENTERPRISE"]'::jsonb, false, false),
('unlimited_users', 'No user limit restrictions', '["ENTERPRISE"]'::jsonb, false, true)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    plan_requirements = EXCLUDED.plan_requirements,
    is_trial_allowed = EXCLUDED.is_trial_allowed,
    is_enterprise_only = EXCLUDED.is_enterprise_only,
    updated_at = CURRENT_TIMESTAMP;

-- ============================================================================
-- STORED PROCEDURES FOR SUBSCRIPTION MANAGEMENT
-- ============================================================================

-- Function to start organization trial (14 days)
CREATE OR REPLACE FUNCTION start_organization_trial(org_id VARCHAR(255))
RETURNS VOID AS $$
DECLARE
    starter_plan_id UUID;
    trial_start TIMESTAMP := CURRENT_TIMESTAMP;
    trial_end TIMESTAMP := CURRENT_TIMESTAMP + INTERVAL '14 days';
BEGIN
    -- Get STARTER_PLAN ID
    SELECT id INTO starter_plan_id FROM subscription_plans WHERE slug = 'STARTER_PLAN' AND is_active = true;
    
    IF starter_plan_id IS NULL THEN
        RAISE EXCEPTION 'STARTER_PLAN not found or inactive';
    END IF;
    
    -- Update organization with trial information
    UPDATE organizations SET
        trial_start_date = trial_start,
        trial_end_date = trial_end,
        current_plan_id = starter_plan_id,
        subscription_status = 'trial',
        max_users_allowed = 50,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = org_id;
    
    -- Create organization subscription record
    INSERT INTO organization_subscriptions (
        organization_id,
        plan_id,
        status,
        current_period_start,
        current_period_end
    ) VALUES (
        org_id,
        starter_plan_id,
        'trial',
        trial_start,
        trial_end
    ) ON CONFLICT (organization_id) DO UPDATE SET
        plan_id = EXCLUDED.plan_id,
        status = EXCLUDED.status,
        current_period_start = EXCLUDED.current_period_start,
        current_period_end = EXCLUDED.current_period_end,
        updated_at = CURRENT_TIMESTAMP;
    
    -- Create audit log
    INSERT INTO subscription_audit_logs (
        organization_id,
        action,
        new_plan_id,
        new_status,
        performed_by,
        metadata
    ) VALUES (
        org_id,
        'trial_started',
        starter_plan_id,
        'trial',
        'system',
        jsonb_build_object(
            'trial_duration_days', 14,
            'trial_start', trial_start,
            'trial_end', trial_end
        )
    );
END;
$$ LANGUAGE plpgsql;

-- Function to check if organization has access to a feature
CREATE OR REPLACE FUNCTION organization_has_feature(org_id VARCHAR(255), feature_name VARCHAR(100))
RETURNS BOOLEAN AS $$
DECLARE
    org_status VARCHAR(50);
    org_plan_slug VARCHAR(50);
    org_trial_end TIMESTAMP;
    org_grace_end TIMESTAMP;
    feature_plans JSONB;
    is_trial_allowed BOOLEAN;
    current_time TIMESTAMP := CURRENT_TIMESTAMP;
BEGIN
    -- Get organization subscription details
    SELECT 
        o.subscription_status,
        sp.slug,
        o.trial_end_date,
        o.grace_period_ends_at
    INTO org_status, org_plan_slug, org_trial_end, org_grace_end
    FROM organizations o
    LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
    WHERE o.id = org_id;
    
    -- If organization not found, deny access
    IF org_status IS NULL THEN
        RETURN FALSE;
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
                -- Only allow basic features during grace period
                RETURN feature_name IN ('core_workflows', 'document_verification', 'standard_analytics');
            ELSE
                -- Trial expired, no grace period or grace period expired
                RETURN FALSE;
            END IF;
        ELSE
            -- Trial active, check if feature is allowed for trial
            IF is_trial_allowed THEN
                RETURN TRUE;
            ELSE
                -- Check if current plan supports the feature
                RETURN feature_plans ? org_plan_slug;
            END IF;
        END IF;
    END IF;
    
    -- For active subscriptions, check plan requirements
    IF org_status = 'active' THEN
        RETURN feature_plans ? org_plan_slug;
    END IF;
    
    -- For past_due status, allow basic access for 7 days
    IF org_status = 'past_due' THEN
        IF org_grace_end IS NOT NULL AND current_time <= org_grace_end THEN
            -- Only allow basic features during grace period
            RETURN feature_name IN ('core_workflows', 'document_verification', 'standard_analytics');
        ELSE
            RETURN FALSE;
        END IF;
    END IF;
    
    -- For canceled or expired, deny access
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-start trial on organization creation
CREATE OR REPLACE FUNCTION trigger_start_organization_trial()
RETURNS TRIGGER AS $$
BEGIN
    -- Only start trial if this is a new organization and no trial dates are set
    IF TG_OP = 'INSERT' AND NEW.trial_start_date IS NULL THEN
        PERFORM start_organization_trial(NEW.id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
DROP TRIGGER IF EXISTS auto_start_trial ON organizations;
CREATE TRIGGER auto_start_trial
    AFTER INSERT ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION trigger_start_organization_trial();

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_organizations_subscription_status ON organizations(subscription_status);
CREATE INDEX IF NOT EXISTS idx_organizations_trial_end_date ON organizations(trial_end_date);
CREATE INDEX IF NOT EXISTS idx_organizations_current_plan_id ON organizations(current_plan_id);
CREATE INDEX IF NOT EXISTS idx_org_subscriptions_status ON organization_subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_org_subscriptions_org_id ON organization_subscriptions(organization_id);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_slug ON subscription_plans(slug);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_active ON subscription_plans(is_active);
CREATE INDEX IF NOT EXISTS idx_feature_flags_name ON feature_flags(name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_org_id ON subscription_audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_performed_at ON subscription_audit_logs(performed_at);

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- View for organization subscription details
CREATE OR REPLACE VIEW organization_subscription_details AS
SELECT 
    o.id as organization_id,
    o.name as organization_name,
    o.subscription_status,
    o.trial_start_date,
    o.trial_end_date,
    o.grace_period_ends_at,
    o.max_users_allowed,
    sp.name as plan_name,
    sp.slug as plan_slug,
    sp.price_monthly,
    sp.price_yearly,
    sp.features as plan_features,
    sp.max_users as plan_max_users,
    os.stripe_subscription_id,
    os.current_period_start,
    os.current_period_end,
    os.cancel_at_period_end,
    os.payment_failed_count,
    os.last_payment_failed_at,
    CASE 
        WHEN o.subscription_status = 'trial' AND CURRENT_TIMESTAMP <= o.trial_end_date THEN
            EXTRACT(DAYS FROM o.trial_end_date - CURRENT_TIMESTAMP)::INTEGER
        ELSE 0
    END as trial_days_remaining,
    CASE 
        WHEN o.subscription_status = 'trial' AND CURRENT_TIMESTAMP > o.trial_end_date THEN true
        ELSE false
    END as trial_expired,
    CASE 
        WHEN o.grace_period_ends_at IS NOT NULL AND CURRENT_TIMESTAMP <= o.grace_period_ends_at THEN true
        ELSE false
    END as in_grace_period
FROM organizations o
LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
LEFT JOIN organization_subscriptions os ON o.id = os.organization_id;

-- ============================================================================
-- CLEANUP AND FINAL SETUP
-- ============================================================================

-- Update existing organizations to have trial if they don't have subscription info
UPDATE organizations 
SET subscription_status = 'trial'
WHERE subscription_status IS NULL 
   AND trial_start_date IS NULL 
   AND current_plan_id IS NULL;

-- Start trials for existing organizations that don't have them
DO $$
DECLARE
    org_record RECORD;
BEGIN
    FOR org_record IN 
        SELECT id FROM organizations 
        WHERE trial_start_date IS NULL 
          AND current_plan_id IS NULL
    LOOP
        PERFORM start_organization_trial(org_record.id);
    END LOOP;
END $$;