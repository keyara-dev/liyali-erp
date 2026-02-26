-- Migration: 015_per_document_type_limits
-- Description: Add per-document-type resource limits to subscription tiers
-- Date: 2026-02-26

-- ============================================================================
-- STEP 1: Add new limit columns to subscription_tiers
-- ============================================================================

ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_requisitions INTEGER NOT NULL DEFAULT 100;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_budgets INTEGER NOT NULL DEFAULT 20;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_purchase_orders INTEGER NOT NULL DEFAULT 50;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_payment_vouchers INTEGER NOT NULL DEFAULT 50;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_grns INTEGER NOT NULL DEFAULT 50;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_departments INTEGER NOT NULL DEFAULT 5;
ALTER TABLE subscription_tiers ADD COLUMN IF NOT EXISTS max_vendors INTEGER NOT NULL DEFAULT 50;

-- ============================================================================
-- STEP 2: Add new limit override columns to organization_limit_overrides
-- ============================================================================

ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_requisitions INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_budgets INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_purchase_orders INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_payment_vouchers INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_grns INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_departments INTEGER;
ALTER TABLE organization_limit_overrides ADD COLUMN IF NOT EXISTS max_vendors INTEGER;

-- ============================================================================
-- STEP 3: Update tier limits with per-document-type values
-- ============================================================================

-- STARTER tier
UPDATE subscription_tiers SET
    max_requisitions = 100,
    max_budgets = 20,
    max_purchase_orders = 50,
    max_payment_vouchers = 50,
    max_grns = 50,
    max_departments = 5,
    max_vendors = 50,
    updated_at = CURRENT_TIMESTAMP
WHERE name = 'starter';

-- PRO tier
UPDATE subscription_tiers SET
    max_requisitions = 500,
    max_budgets = 100,
    max_purchase_orders = 300,
    max_payment_vouchers = 300,
    max_grns = 300,
    max_departments = 25,
    max_vendors = 200,
    updated_at = CURRENT_TIMESTAMP
WHERE name = 'pro';

-- CUSTOM tier (unlimited)
UPDATE subscription_tiers SET
    max_requisitions = -1,
    max_budgets = -1,
    max_purchase_orders = -1,
    max_payment_vouchers = -1,
    max_grns = -1,
    max_departments = -1,
    max_vendors = -1,
    updated_at = CURRENT_TIMESTAMP
WHERE name = 'custom';

-- ============================================================================
-- STEP 4: Audit log entry
-- ============================================================================

INSERT INTO admin_audit_logs (id, action, details, reason, admin_user_id, created_at)
VALUES (
    'audit-migration-015-' || EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::TEXT,
    'system_migration',
    jsonb_build_object(
        'migration', '015_per_document_type_limits',
        'description', 'Added per-document-type resource limits to subscription tiers',
        'new_columns', jsonb_build_array(
            'max_requisitions', 'max_budgets', 'max_purchase_orders',
            'max_payment_vouchers', 'max_grns', 'max_departments', 'max_vendors'
        ),
        'completed_at', CURRENT_TIMESTAMP
    ),
    'Add granular resource limits for each document type instead of shared document limit',
    'system',
    CURRENT_TIMESTAMP
);
