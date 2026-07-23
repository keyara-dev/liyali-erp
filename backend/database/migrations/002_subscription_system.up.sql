-- ============================================================================
-- LIYALI GATEWAY — CONSOLIDATED SUBSCRIPTION SYSTEM
-- Migration: 002_subscription_system
-- Replaces: 007, 008, 011, 014, 015, 027_fix, 028, 029
-- Final schemas baked in — no ALTER TABLE required
-- ============================================================================

-- ============================================================================
-- SUBSCRIPTION PLANS
-- ============================================================================
CREATE TABLE IF NOT EXISTS subscription_plans (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    slug          VARCHAR(50)  UNIQUE NOT NULL,
    description   TEXT,
    price_monthly DECIMAL(10,2) DEFAULT 0.00,
    price_yearly  DECIMAL(10,2) DEFAULT 0.00,
    features      JSONB        NOT NULL DEFAULT '[]',
    max_users     INTEGER      DEFAULT 50,
    is_active     BOOLEAN      DEFAULT true,
    sort_order    INTEGER      DEFAULT 0,
    metadata      JSONB        DEFAULT '{}',
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_subscription_plan_slug
        CHECK (slug IN ('STARTER_PLAN','PRO_PLAN','ENTERPRISE'))
);

-- Deferred FK: organizations.current_plan_id → subscription_plans
-- (Can only be added after subscription_plans exists)
ALTER TABLE organizations
    ADD CONSTRAINT fk_organizations_current_plan
    FOREIGN KEY (current_plan_id) REFERENCES subscription_plans(id);

-- ============================================================================
-- ORGANIZATION SUBSCRIPTIONS
-- ============================================================================
CREATE TABLE IF NOT EXISTS organization_subscriptions (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id        VARCHAR(255) NOT NULL,
    plan_id                UUID         NOT NULL,
    stripe_subscription_id VARCHAR(255),
    status                 VARCHAR(50)  NOT NULL DEFAULT 'trial',
    current_period_start   TIMESTAMP WITH TIME ZONE,
    current_period_end     TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end   BOOLEAN      DEFAULT false,
    payment_failed_count   INTEGER      DEFAULT 0,
    last_payment_failed_at TIMESTAMP WITH TIME ZONE,
    created_at             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_org_subscriptions_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_org_subscriptions_plan
        FOREIGN KEY (plan_id) REFERENCES subscription_plans(id),
    CONSTRAINT check_subscription_status
        CHECK (status IN ('trial','active','past_due','canceled','expired')),
    CONSTRAINT uk_org_subscription UNIQUE (organization_id)
);

-- ============================================================================
-- SUBSCRIPTION AUDIT LOGS
-- ============================================================================
CREATE TABLE IF NOT EXISTS subscription_audit_logs (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    action       VARCHAR(100) NOT NULL,
    old_plan_id  UUID,
    new_plan_id  UUID,
    old_status   VARCHAR(50),
    new_status   VARCHAR(50),
    metadata     JSONB        DEFAULT '{}',
    performed_by VARCHAR(255),
    performed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_audit_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_audit_old_plan
        FOREIGN KEY (old_plan_id) REFERENCES subscription_plans(id),
    CONSTRAINT fk_audit_new_plan
        FOREIGN KEY (new_plan_id) REFERENCES subscription_plans(id)
);

-- ============================================================================
-- SUBSCRIPTION TIERS  (final schema: 011 + 014 rename + 015 doc-type columns)
-- ============================================================================
CREATE TABLE IF NOT EXISTS subscription_tiers (
    id                  VARCHAR(255) PRIMARY KEY,
    name                VARCHAR(255) UNIQUE NOT NULL,
    display_name        VARCHAR(255) NOT NULL,
    description         TEXT         NOT NULL,
    price_monthly       NUMERIC(10,2) NOT NULL DEFAULT 0,
    price_yearly        NUMERIC(10,2) NOT NULL DEFAULT 0,
    -- limits (max_team_members = renamed from max_users in 014)
    max_workspaces      INTEGER      NOT NULL DEFAULT 1,
    max_team_members    INTEGER      NOT NULL DEFAULT 1,
    max_documents       INTEGER      NOT NULL DEFAULT 100,
    max_workflows       INTEGER      NOT NULL DEFAULT 1,
    max_custom_roles    INTEGER      NOT NULL DEFAULT 0,
    max_requisitions    INTEGER      NOT NULL DEFAULT 100,
    max_budgets         INTEGER      NOT NULL DEFAULT 20,
    max_purchase_orders INTEGER      NOT NULL DEFAULT 50,
    max_payment_vouchers INTEGER     NOT NULL DEFAULT 50,
    max_grns            INTEGER      NOT NULL DEFAULT 50,
    max_departments     INTEGER      NOT NULL DEFAULT 5,
    max_vendors         INTEGER      NOT NULL DEFAULT 50,
    features            JSONB        NOT NULL DEFAULT '[]',
    is_active           BOOLEAN      NOT NULL DEFAULT true,
    sort_order          INTEGER      NOT NULL DEFAULT 0,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- SUBSCRIPTION FEATURES
-- ============================================================================
CREATE TABLE IF NOT EXISTS subscription_features (
    id           VARCHAR(255) PRIMARY KEY,
    name         VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description  TEXT         NOT NULL,
    category     VARCHAR(100) NOT NULL,
    is_active    BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- ORGANIZATION LIMIT OVERRIDES  (final schema: 011 + 014 rename + 015 columns)
-- ============================================================================
CREATE TABLE IF NOT EXISTS organization_limit_overrides (
    id                   VARCHAR(255) PRIMARY KEY,
    organization_id      VARCHAR(255) NOT NULL,
    max_team_members     INTEGER,
    max_workspaces       INTEGER,
    max_documents        INTEGER,
    max_workflows        INTEGER,
    max_custom_roles     INTEGER,
    max_requisitions     INTEGER,
    max_budgets          INTEGER,
    max_purchase_orders  INTEGER,
    max_payment_vouchers INTEGER,
    max_grns             INTEGER,
    max_departments      INTEGER,
    max_vendors          INTEGER,
    features             JSONB,
    reason               TEXT         NOT NULL,
    admin_user_id        VARCHAR(255) NOT NULL,
    expires_at           TIMESTAMP WITH TIME ZONE,
    created_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_org_limit_overrides_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- ============================================================================
-- ADMIN AUDIT LOGS  (subscription/tier changes)
-- ============================================================================
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id              VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255),
    action          VARCHAR(255) NOT NULL,
    old_value       TEXT,
    new_value       TEXT,
    details         JSONB,
    reason          TEXT,
    admin_user_id   VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_audit_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- ============================================================================
-- BILLING TABLES  (012)
-- ============================================================================
CREATE TABLE IF NOT EXISTS payments (
    id                   VARCHAR(255) PRIMARY KEY,
    organization_id      VARCHAR(255) NOT NULL,
    subscription_tier    VARCHAR(100) NOT NULL,
    amount               NUMERIC(10,2) NOT NULL,
    currency             VARCHAR(10)  NOT NULL DEFAULT 'USD',
    payment_method       VARCHAR(50),
    payment_status       VARCHAR(50)  NOT NULL,
    billing_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    billing_period_end   TIMESTAMP WITH TIME ZONE NOT NULL,
    invoice_id           VARCHAR(255),
    transaction_id       VARCHAR(255),
    metadata             JSONB,
    paid_at              TIMESTAMP WITH TIME ZONE,
    created_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_payments_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS invoices (
    id              VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    invoice_number  VARCHAR(100) UNIQUE NOT NULL,
    amount          NUMERIC(10,2) NOT NULL,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'USD',
    status          VARCHAR(50)  NOT NULL,
    due_date        TIMESTAMP WITH TIME ZONE NOT NULL,
    paid_date       TIMESTAMP WITH TIME ZONE,
    items           JSONB        NOT NULL,
    metadata        JSONB,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_invoices_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS subscription_events (
    id              VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    from_tier       VARCHAR(100),
    to_tier         VARCHAR(100),
    metadata        JSONB,
    created_by      VARCHAR(255),
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_subscription_events_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- ============================================================================
-- SUBSCRIPTION FEATURE REQUIREMENTS  (029 — used by organization_has_feature())
-- ============================================================================
CREATE TABLE IF NOT EXISTS subscription_feature_requirements (
    name               VARCHAR(100) PRIMARY KEY,
    description        TEXT,
    plan_requirements  JSONB        NOT NULL DEFAULT '[]',
    is_trial_allowed   BOOLEAN      NOT NULL DEFAULT FALSE,
    is_enterprise_only BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- INDEXES
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_subscription_plans_slug         ON subscription_plans(slug);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_active       ON subscription_plans(is_active);
CREATE INDEX IF NOT EXISTS idx_org_subscriptions_status        ON organization_subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_org_subscriptions_org_id        ON organization_subscriptions(organization_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_org_id               ON subscription_audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_performed_at         ON subscription_audit_logs(performed_at);
CREATE INDEX IF NOT EXISTS idx_subscription_tiers_name         ON subscription_tiers(name);
CREATE INDEX IF NOT EXISTS idx_subscription_tiers_active       ON subscription_tiers(is_active);
CREATE INDEX IF NOT EXISTS idx_subscription_tiers_sort         ON subscription_tiers(sort_order);
CREATE INDEX IF NOT EXISTS idx_subscription_features_name      ON subscription_features(name);
CREATE INDEX IF NOT EXISTS idx_subscription_features_category  ON subscription_features(category);
CREATE INDEX IF NOT EXISTS idx_subscription_features_active    ON subscription_features(is_active);
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_overrides_unique ON organization_limit_overrides(organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_overrides_org      ON organization_limit_overrides(organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_overrides_expires  ON organization_limit_overrides(expires_at);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_org            ON admin_audit_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_action         ON admin_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_created        ON admin_audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_payments_org                    ON payments(organization_id);
CREATE INDEX IF NOT EXISTS idx_payments_status                 ON payments(payment_status);
CREATE INDEX IF NOT EXISTS idx_payments_created                ON payments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_period                 ON payments(billing_period_start, billing_period_end);
CREATE INDEX IF NOT EXISTS idx_invoices_org                    ON invoices(organization_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status                 ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_due                    ON invoices(due_date);
CREATE INDEX IF NOT EXISTS idx_subscription_events_org         ON subscription_events(organization_id);
CREATE INDEX IF NOT EXISTS idx_subscription_events_type        ON subscription_events(event_type);
CREATE INDEX IF NOT EXISTS idx_subscription_events_created     ON subscription_events(created_at DESC);

-- ============================================================================
-- TRIGGERS (updated_at)
-- ============================================================================
DROP TRIGGER IF EXISTS update_subscription_tiers_updated_at ON subscription_tiers;
CREATE TRIGGER update_subscription_tiers_updated_at
    BEFORE UPDATE ON subscription_tiers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organization_overrides_updated_at ON organization_limit_overrides;
CREATE TRIGGER update_organization_overrides_updated_at
    BEFORE UPDATE ON organization_limit_overrides FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
CREATE TRIGGER update_invoices_updated_at
    BEFORE UPDATE ON invoices FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- STORED FUNCTIONS
-- ============================================================================

-- Start a 14-day trial for an organization
CREATE OR REPLACE FUNCTION start_organization_trial(org_id VARCHAR(255))
RETURNS VOID AS $$
DECLARE
    starter_plan_id UUID;
    trial_start TIMESTAMP := CURRENT_TIMESTAMP;
    trial_end   TIMESTAMP := CURRENT_TIMESTAMP + INTERVAL '14 days';
BEGIN
    SELECT id INTO starter_plan_id
    FROM subscription_plans
    WHERE slug = 'STARTER_PLAN' AND is_active = true;

    IF starter_plan_id IS NULL THEN
        RAISE EXCEPTION 'STARTER_PLAN not found or inactive';
    END IF;

    UPDATE organizations SET
        trial_start_date  = trial_start,
        trial_end_date    = trial_end,
        current_plan_id   = starter_plan_id,
        subscription_status = 'trial',
        max_users_allowed = 50,
        updated_at        = CURRENT_TIMESTAMP
    WHERE id = org_id;

    INSERT INTO organization_subscriptions (
        organization_id, plan_id, status,
        current_period_start, current_period_end
    ) VALUES (
        org_id, starter_plan_id, 'trial', trial_start, trial_end
    ) ON CONFLICT (organization_id) DO UPDATE SET
        plan_id              = EXCLUDED.plan_id,
        status               = EXCLUDED.status,
        current_period_start = EXCLUDED.current_period_start,
        current_period_end   = EXCLUDED.current_period_end,
        updated_at           = CURRENT_TIMESTAMP;

    INSERT INTO subscription_audit_logs (
        organization_id, action, new_plan_id, new_status, performed_by, metadata
    ) VALUES (
        org_id, 'trial_started', starter_plan_id, 'trial', 'system',
        jsonb_build_object(
            'trial_duration_days', 14,
            'trial_start', trial_start,
            'trial_end',   trial_end
        )
    );
END;
$$ LANGUAGE plpgsql;

-- Auto-start trial on org INSERT (only when trial_start_date IS NULL)
CREATE OR REPLACE FUNCTION trigger_start_organization_trial()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.trial_start_date IS NULL THEN
        PERFORM start_organization_trial(NEW.id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS auto_start_trial ON organizations;
CREATE TRIGGER auto_start_trial
    AFTER INSERT ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION trigger_start_organization_trial();

-- Check if organization has access to a feature
-- Uses subscription_feature_requirements (not the admin-UI feature_flags table)
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
    IF org_tier IN ('pro', 'enterprise') THEN
        IF org_plan_slug IS NULL THEN
            org_plan_slug := CASE org_tier
                WHEN 'pro'    THEN 'PRO_PLAN'
                WHEN 'enterprise' THEN 'ENTERPRISE'
                ELSE org_tier
            END;
        END IF;
        SELECT plan_requirements INTO feature_plans
        FROM subscription_feature_requirements
        WHERE name = feature_name;
        IF feature_plans IS NULL THEN
            RETURN FALSE;
        END IF;
        RETURN feature_plans ? org_plan_slug
            OR feature_plans ? 'PRO_PLAN'
            OR feature_plans ? 'ENTERPRISE';
    END IF;

    SELECT plan_requirements, is_trial_allowed
    INTO feature_plans, is_trial_ok
    FROM subscription_feature_requirements
    WHERE name = feature_name;

    IF feature_plans IS NULL THEN
        RETURN FALSE;
    END IF;

    IF org_status = 'trial' THEN
        IF org_trial_end IS NOT NULL AND current_time > org_trial_end THEN
            IF org_grace_end IS NOT NULL AND current_time <= org_grace_end THEN
                RETURN feature_name IN ('core_workflows','document_verification','standard_analytics');
            ELSE
                RETURN FALSE;
            END IF;
        ELSE
            IF is_trial_ok THEN
                RETURN TRUE;
            ELSE
                RETURN feature_plans ? COALESCE(org_plan_slug, 'STARTER_PLAN');
            END IF;
        END IF;
    END IF;

    IF org_status = 'active' THEN
        RETURN feature_plans ? COALESCE(org_plan_slug, 'STARTER_PLAN');
    END IF;

    IF org_status = 'past_due' THEN
        IF org_grace_end IS NOT NULL AND current_time <= org_grace_end THEN
            RETURN feature_name IN ('core_workflows','document_verification','standard_analytics');
        ELSE
            RETURN FALSE;
        END IF;
    END IF;

    RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Track subscription tier changes
CREATE OR REPLACE FUNCTION track_subscription_tier_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.subscription_tier IS DISTINCT FROM NEW.subscription_tier THEN
        INSERT INTO subscription_events (id, organization_id, event_type, from_tier, to_tier, created_at)
        VALUES (
            'event-' || gen_random_uuid()::text,
            NEW.id,
            CASE
                WHEN NEW.subscription_tier > OLD.subscription_tier THEN 'subscription_upgraded'
                ELSE 'subscription_downgraded'
            END,
            OLD.subscription_tier,
            NEW.subscription_tier,
            CURRENT_TIMESTAMP
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS track_subscription_tier_changes ON organizations;
CREATE TRIGGER track_subscription_tier_changes
    AFTER UPDATE OF subscription_tier ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION track_subscription_tier_changes();

-- Track trial conversions
CREATE OR REPLACE FUNCTION track_trial_conversions()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.subscription_status = 'trial' AND NEW.subscription_status = 'active' THEN
        INSERT INTO subscription_events (id, organization_id, event_type, from_tier, to_tier, created_at)
        VALUES (
            'event-' || gen_random_uuid()::text,
            NEW.id,
            'trial_converted',
            OLD.subscription_tier,
            NEW.subscription_tier,
            CURRENT_TIMESTAMP
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS track_trial_conversions ON organizations;
CREATE TRIGGER track_trial_conversions
    AFTER UPDATE OF subscription_status ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION track_trial_conversions();

-- ============================================================================
-- VIEW: organization subscription details
-- ============================================================================
CREATE OR REPLACE VIEW organization_subscription_details AS
SELECT
    o.id                    AS organization_id,
    o.name                  AS organization_name,
    o.subscription_status,
    o.trial_start_date,
    o.trial_end_date,
    o.grace_period_ends_at,
    o.max_users_allowed,
    sp.name                 AS plan_name,
    sp.slug                 AS plan_slug,
    sp.price_monthly,
    sp.price_yearly,
    sp.features             AS plan_features,
    sp.max_users            AS plan_max_users,
    os.stripe_subscription_id,
    os.current_period_start,
    os.current_period_end,
    os.cancel_at_period_end,
    os.payment_failed_count,
    os.last_payment_failed_at,
    CASE
        WHEN o.subscription_status = 'trial' AND CURRENT_TIMESTAMP <= o.trial_end_date
            THEN EXTRACT(DAYS FROM o.trial_end_date - CURRENT_TIMESTAMP)::INTEGER
        ELSE 0
    END AS trial_days_remaining,
    CASE
        WHEN o.subscription_status = 'trial' AND CURRENT_TIMESTAMP > o.trial_end_date
            THEN true ELSE false
    END AS trial_expired,
    CASE
        WHEN o.grace_period_ends_at IS NOT NULL AND CURRENT_TIMESTAMP <= o.grace_period_ends_at
            THEN true ELSE false
    END AS in_grace_period
FROM organizations o
LEFT JOIN subscription_plans sp ON o.current_plan_id = sp.id
LEFT JOIN organization_subscriptions os ON o.id = os.organization_id;

-- ============================================================================
-- SEED: Subscription plans
-- ============================================================================
INSERT INTO subscription_plans (name, slug, description, price_monthly, price_yearly, features, max_users, sort_order, metadata)
VALUES
(
    'Starter Plan', 'STARTER_PLAN',
    'Perfect for small teams getting started with procurement workflows',
    0.00, 0.00,
    '["Core procurement workflows","Up to 50 users","Single Workspace","Document Verification (QR Codes and Doc Numbers)","Standard analytics","Notifications (Email & In-App)"]'::jsonb,
    50, 1,
    '{"offline_capabilities":false,"api_access":false,"custom_roles":false,"priority_support":false,"dedicated_instance":false,"sla_guarantees":false}'::jsonb
),
(
    'Pro Plan', 'PRO_PLAN',
    'Advanced features for growing organizations',
    99.00, 990.00,
    '["Everything in Starter Plan","Up to 200 users","Custom Role management","Offline capabilities","Priority support","Advanced analytics","API Access"]'::jsonb,
    200, 2,
    '{"offline_capabilities":true,"api_access":true,"custom_roles":true,"priority_support":true,"dedicated_instance":false,"sla_guarantees":false}'::jsonb
),
(
    'Enterprise', 'ENTERPRISE',
    'Complete solution for large organizations',
    0.00, 0.00,
    '["Everything in Pro Plan","Unlimited users","Dedicated instance","Custom integrations","SLA guarantees","Dedicated success manager","Models Creation/Modifications"]'::jsonb,
    -1, 3,
    '{"offline_capabilities":true,"api_access":true,"custom_roles":true,"priority_support":true,"dedicated_instance":true,"sla_guarantees":true,"custom_pricing":true}'::jsonb
)
ON CONFLICT (slug) DO UPDATE SET
    name          = EXCLUDED.name,
    description   = EXCLUDED.description,
    price_monthly = EXCLUDED.price_monthly,
    price_yearly  = EXCLUDED.price_yearly,
    features      = EXCLUDED.features,
    max_users     = EXCLUDED.max_users,
    sort_order    = EXCLUDED.sort_order,
    metadata      = EXCLUDED.metadata,
    updated_at    = CURRENT_TIMESTAMP;

-- ============================================================================
-- SEED: Subscription tiers (final values from 014 + 015)
-- ============================================================================
INSERT INTO subscription_tiers (
    id, name, display_name, description,
    price_monthly, price_yearly,
    max_workspaces, max_team_members, max_documents, max_workflows, max_custom_roles,
    max_requisitions, max_budgets, max_purchase_orders, max_payment_vouchers,
    max_grns, max_departments, max_vendors,
    features, is_active, sort_order
) VALUES
(
    'tier-starter', 'starter', 'Starter', 'Perfect for small teams getting started',
    0, 0,
    1, 10, 200, 3, 0,
    100, 20, 50, 50, 50, 5, 50,
    '["document_management","basic_workflows","in_app_notifications","standard_reports","user_management","department_management","vendor_management","budget_tracking","mobile_web_access","email_support"]'::jsonb,
    true, 1
),
(
    'tier-pro', 'pro', 'Pro', 'Advanced features for growing organizations',
    99, 990,
    5, 50, 500, 20, 10,
    500, 100, 300, 300, 300, 25, 200,
    '["document_management","basic_workflows","in_app_notifications","standard_reports","user_management","department_management","vendor_management","budget_tracking","mobile_web_access","email_support","advanced_workflows","email_notifications","custom_roles","advanced_analytics","data_export","priority_support","audit_logs_90_days","multi_currency","advanced_reporting","workflow_templates"]'::jsonb,
    true, 2
),
(
    'tier-enterprise', 'enterprise', 'Enterprise', 'Enterprise solution with unlimited configurable limits',
    499, 4990,
    -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1,
    '["document_management","basic_workflows","in_app_notifications","standard_reports","user_management","department_management","vendor_management","budget_tracking","mobile_web_access","email_support","advanced_workflows","email_notifications","custom_roles","advanced_analytics","data_export","priority_support","audit_logs_90_days","multi_currency","advanced_reporting","workflow_templates","webhooks","custom_fields","bulk_operations","api_access","dedicated_support_manager","sla_guarantees","audit_logs_unlimited","custom_development","professional_services","dedicated_support","custom_training"]'::jsonb,
    true, 3
)
ON CONFLICT (id) DO UPDATE SET
    name                 = EXCLUDED.name,
    display_name         = EXCLUDED.display_name,
    description          = EXCLUDED.description,
    price_monthly        = EXCLUDED.price_monthly,
    price_yearly         = EXCLUDED.price_yearly,
    max_workspaces       = EXCLUDED.max_workspaces,
    max_team_members     = EXCLUDED.max_team_members,
    max_documents        = EXCLUDED.max_documents,
    max_workflows        = EXCLUDED.max_workflows,
    max_custom_roles     = EXCLUDED.max_custom_roles,
    max_requisitions     = EXCLUDED.max_requisitions,
    max_budgets          = EXCLUDED.max_budgets,
    max_purchase_orders  = EXCLUDED.max_purchase_orders,
    max_payment_vouchers = EXCLUDED.max_payment_vouchers,
    max_grns             = EXCLUDED.max_grns,
    max_departments      = EXCLUDED.max_departments,
    max_vendors          = EXCLUDED.max_vendors,
    features             = EXCLUDED.features,
    is_active            = EXCLUDED.is_active,
    sort_order           = EXCLUDED.sort_order,
    updated_at           = CURRENT_TIMESTAMP;

-- ============================================================================
-- SEED: Subscription features (031 final 31-row set from 014)
-- ============================================================================

-- STARTER features
INSERT INTO subscription_features (id, name, display_name, description, category, is_active) VALUES
('feat-doc-mgmt',        'document_management',    'Document Management',    'Create, edit, and manage documents',                'core',     true),
('feat-basic-workflows', 'basic_workflows',         'Basic Workflows',         'Simple linear approval workflows (1-3 stages)',      'workflow',  true),
('feat-in-app-notif',    'in_app_notifications',    'In-App Notifications',   'Real-time in-app notifications',                    'core',     true),
('feat-std-reports',     'standard_reports',        'Standard Reports',        'Pre-built standard reports',                        'analytics', true),
('feat-user-mgmt',       'user_management',         'User Management',         'Manage users with system roles',                    'core',     true),
('feat-dept-mgmt',       'department_management',   'Department Management',  'Organize users into departments',                   'core',     true),
('feat-vendor-mgmt',     'vendor_management',       'Vendor Management',      'Maintain vendor database',                          'core',     true),
('feat-budget-track',    'budget_tracking',         'Budget Tracking',        'Track budget utilization',                          'core',     true),
('feat-mobile-web',      'mobile_web_access',       'Mobile Web Access',      'Responsive mobile web interface',                   'core',     true),
('feat-email-support',   'email_support',           'Email Support',          'Standard email support',                            'support',  true)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name, display_name = EXCLUDED.display_name,
    description = EXCLUDED.description, category = EXCLUDED.category, is_active = EXCLUDED.is_active;

-- PRO features
INSERT INTO subscription_features (id, name, display_name, description, category, is_active) VALUES
('feat-adv-workflows',     'advanced_workflows',    'Advanced Workflows',            'Complex multi-stage workflows with conditional routing', 'workflow',  true),
('feat-email-notif',       'email_notifications',   'Email Notifications',           'Automated email notifications',                          'core',     true),
('feat-custom-roles',      'custom_roles',          'Custom Roles & Permissions',    'Create organization-specific roles',                     'security', true),
('feat-adv-analytics',     'advanced_analytics',    'Advanced Analytics & Dashboards','Custom dashboards and analytics',                       'analytics', true),
('feat-data-export',       'data_export',           'Data Export',                   'Export data to CSV/Excel',                               'analytics', true),
('feat-priority-support',  'priority_support',      'Priority Support',              '24/7 priority customer support',                         'support',  true),
('feat-audit-90',          'audit_logs_90_days',    'Audit Logs (90 days)',          'Audit trail with 90-day retention',                      'security', true),
('feat-multi-currency',    'multi_currency',        'Multi-Currency Support',        'Support for multiple currencies',                        'core',     true),
('feat-adv-reporting',     'advanced_reporting',    'Advanced Reporting',            'Custom report builder',                                  'analytics', true),
('feat-workflow-templates','workflow_templates',    'Workflow Templates',            'Pre-built workflow templates',                           'workflow',  true)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name, display_name = EXCLUDED.display_name,
    description = EXCLUDED.description, category = EXCLUDED.category, is_active = EXCLUDED.is_active;

-- CUSTOM features
INSERT INTO subscription_features (id, name, display_name, description, category, is_active) VALUES
('feat-webhooks',         'webhooks',                 'Webhooks',                'Real-time event webhooks',                  'integration',   true),
('feat-custom-fields',    'custom_fields',            'Custom Fields',           'Add custom fields to documents',            'customization', true),
('feat-bulk-ops',         'bulk_operations',          'Bulk Operations',         'Batch process documents',                   'core',          true),
('feat-api-access',       'api_access',               'API Access',              'Full REST API access',                      'integration',   true),
('feat-dedicated-mgr',    'dedicated_support_manager','Dedicated Support Manager','Dedicated customer success manager',       'support',       true),
('feat-sla',              'sla_guarantees',           'SLA Guarantees',          'Service level agreement guarantees',        'support',       true),
('feat-audit-unlimited',  'audit_logs_unlimited',     'Audit Logs (Unlimited)',  'Unlimited audit log retention',             'security',      true),
('feat-custom-dev',       'custom_development',       'Custom Development',      'Bespoke feature development',               'customization', true),
('feat-prof-services',    'professional_services',    'Professional Services',   'Implementation consulting',                 'support',       true),
('feat-dedicated-support','dedicated_support',        'Dedicated Support',       '24/7 dedicated support team',               'support',       true),
('feat-custom-training',  'custom_training',          'Custom Training',         'Tailored training sessions',                'support',       true)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name, display_name = EXCLUDED.display_name,
    description = EXCLUDED.description, category = EXCLUDED.category, is_active = EXCLUDED.is_active;

-- ============================================================================
-- SEED: Subscription feature requirements (029 — 23 rows)
-- ============================================================================
INSERT INTO subscription_feature_requirements (name, description, plan_requirements, is_trial_allowed, is_enterprise_only) VALUES
-- Starter (free during trial)
('core_workflows',         'Basic approval workflows',                  '["STARTER_PLAN","PRO_PLAN","ENTERPRISE"]'::jsonb, true,  false),
('document_verification',  'Document authenticity verification',        '["STARTER_PLAN","PRO_PLAN","ENTERPRISE"]'::jsonb, true,  false),
('standard_analytics',     'Standard reporting',                        '["STARTER_PLAN","PRO_PLAN","ENTERPRISE"]'::jsonb, true,  false),
-- Pro tier
('custom_roles',           'Create and manage custom user roles',       '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('audit_logs_90_days',     'Audit trail with 90-day retention',         '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('advanced_workflows',     'Complex multi-stage conditional workflows', '["PRO_PLAN","ENTERPRISE"]'::jsonb,             true,  false),
('email_notifications',    'Automated email notifications',             '["PRO_PLAN","ENTERPRISE"]'::jsonb,             true,  false),
('advanced_analytics',     'Advanced reporting and analytics',          '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('data_export',            'Export data to CSV/Excel',                  '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('priority_support',       'Priority customer support',                 '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('multi_currency',         'Support for multiple currencies',           '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('advanced_reporting',     'Custom report builder',                     '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('workflow_templates',     'Pre-built workflow templates',              '["PRO_PLAN","ENTERPRISE"]'::jsonb,             true,  false),
('bulk_operations',        'Batch process documents',                   '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('offline_capabilities',   'Work offline and sync when connected',      '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
('api_access',             'Access to REST API endpoints',              '["PRO_PLAN","ENTERPRISE"]'::jsonb,             false, false),
-- Enterprise only
('webhooks',               'Real-time event webhooks',                  '["ENTERPRISE"]'::jsonb,                        false, false),
('custom_fields',          'Add custom fields to documents',            '["ENTERPRISE"]'::jsonb,                        false, false),
('dedicated_instance',     'Dedicated server instance',                 '["ENTERPRISE"]'::jsonb,                        false, true),
('sla_guarantees',         'Service Level Agreement guarantees',        '["ENTERPRISE"]'::jsonb,                        false, true),
('custom_integrations',    'Custom third-party integrations',           '["ENTERPRISE"]'::jsonb,                        false, true),
('models_modification',    'Create and modify data models',             '["ENTERPRISE"]'::jsonb,                        false, true),
('unlimited_users',        'No user limit restrictions',                '["ENTERPRISE"]'::jsonb,                        false, true),
('audit_logs_unlimited',   'Unlimited audit log retention',             '["ENTERPRISE"]'::jsonb,                        false, true)
ON CONFLICT (name) DO UPDATE SET
    description        = EXCLUDED.description,
    plan_requirements  = EXCLUDED.plan_requirements,
    is_trial_allowed   = EXCLUDED.is_trial_allowed,
    is_enterprise_only = EXCLUDED.is_enterprise_only,
    updated_at         = CURRENT_TIMESTAMP;
