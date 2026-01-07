-- ============================================================================
-- LIYALI GATEWAY - ROLLBACK ORGANIZATION TIER STANDARDIZATION
-- Migration: 003_standardize_organization_tiers (DOWN)
-- Description: Rollback tier standardization changes
-- Date: 2025-01-07
-- ============================================================================

-- Remove the check constraint
ALTER TABLE organizations 
DROP CONSTRAINT IF EXISTS check_organization_tier;

-- Revert tier values back to old system
-- NEW SYSTEM: starter, pro, enterprise
-- OLD SYSTEM: free, premium, pro, enterprise

-- Revert 'starter' back to 'free'
UPDATE organizations 
SET tier = 'free', updated_at = CURRENT_TIMESTAMP 
WHERE tier = 'starter';

-- Note: We'll keep 'pro' as 'pro' and 'enterprise' as 'enterprise'
-- since those were already in the old system

-- Revert the default value
ALTER TABLE organizations 
ALTER COLUMN tier SET DEFAULT 'free';

-- Log the rollback
DO $
DECLARE
    free_count INTEGER;
    pro_count INTEGER;
    enterprise_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO free_count FROM organizations WHERE tier = 'free';
    SELECT COUNT(*) INTO pro_count FROM organizations WHERE tier = 'pro';
    SELECT COUNT(*) INTO enterprise_count FROM organizations WHERE tier = 'enterprise';
    
    RAISE NOTICE 'Migration 003_standardize_organization_tiers rollback completed';
    RAISE NOTICE 'Organizations with FREE tier: %', free_count;
    RAISE NOTICE 'Organizations with PRO tier: %', pro_count;
    RAISE NOTICE 'Organizations with ENTERPRISE tier: %', enterprise_count;
    RAISE NOTICE 'Tier rollback complete';
END $;