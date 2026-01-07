-- ============================================================================
-- LIYALI GATEWAY - STANDARDIZE ORGANIZATION TIERS
-- Migration: 003_standardize_organization_tiers
-- Description: Standardizes organization tier values to match frontend expectations
-- Date: 2025-01-07
-- ============================================================================

-- Update existing tier values to match the standardized system
-- OLD SYSTEM: free, premium, pro, enterprise
-- NEW SYSTEM: starter, pro, enterprise

-- Update 'free' to 'starter'
UPDATE organizations 
SET tier = 'starter', updated_at = CURRENT_TIMESTAMP 
WHERE tier = 'free';

-- Update 'premium' to 'pro'
UPDATE organizations 
SET tier = 'pro', updated_at = CURRENT_TIMESTAMP 
WHERE tier = 'premium';

-- Ensure 'pro' remains 'pro' (no change needed)
-- Ensure 'enterprise' remains 'enterprise' (no change needed)

-- Update the default value for new organizations
ALTER TABLE organizations 
ALTER COLUMN tier SET DEFAULT 'starter';

-- Add a check constraint to ensure only valid tier values
ALTER TABLE organizations 
ADD CONSTRAINT check_organization_tier 
CHECK (tier IN ('starter', 'pro', 'enterprise'));

-- Log the changes
DO $
DECLARE
    starter_count INTEGER;
    pro_count INTEGER;
    enterprise_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO starter_count FROM organizations WHERE tier = 'starter';
    SELECT COUNT(*) INTO pro_count FROM organizations WHERE tier = 'pro';
    SELECT COUNT(*) INTO enterprise_count FROM organizations WHERE tier = 'enterprise';
    
    RAISE NOTICE 'Migration 003_standardize_organization_tiers completed successfully';
    RAISE NOTICE 'Organizations with STARTER tier: %', starter_count;
    RAISE NOTICE 'Organizations with PRO tier: %', pro_count;
    RAISE NOTICE 'Organizations with ENTERPRISE tier: %', enterprise_count;
    RAISE NOTICE 'Tier standardization complete';
END $;