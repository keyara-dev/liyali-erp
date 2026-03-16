-- Migration 028: Fix orgs where the legacy `tier` column is pro/custom
-- but subscription_status was not cleared (migration 027 only checked subscription_tier)

UPDATE organizations
SET subscription_status    = 'active',
    subscription_tier      = tier,
    trial_end_date         = NULL,
    grace_period_ends_at   = NULL,
    updated_at             = CURRENT_TIMESTAMP
WHERE tier IN ('pro', 'custom')
  AND (
    subscription_status != 'active'
    OR subscription_status IS NULL
    OR trial_end_date IS NOT NULL
    OR grace_period_ends_at IS NOT NULL
  );
