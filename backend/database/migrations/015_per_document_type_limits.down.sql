-- Rollback: 015_per_document_type_limits

ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_requisitions;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_budgets;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_purchase_orders;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_payment_vouchers;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_grns;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_departments;
ALTER TABLE subscription_tiers DROP COLUMN IF EXISTS max_vendors;

ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_requisitions;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_budgets;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_purchase_orders;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_payment_vouchers;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_grns;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_departments;
ALTER TABLE organization_limit_overrides DROP COLUMN IF EXISTS max_vendors;
