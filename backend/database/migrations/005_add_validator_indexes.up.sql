-- ============================================================================
-- LIYALI GATEWAY - ADD VALIDATOR INDEXES
-- Migration: 005_add_validator_indexes
-- Description: Add indexes checked by the bootstrap validator
-- Date: February 1, 2026
-- ============================================================================

-- Organizations active index (checked by validator)
CREATE INDEX IF NOT EXISTS idx_organizations_active ON organizations(active);

-- Requisitions organization_id index (checked by validator)
CREATE INDEX IF NOT EXISTS idx_requisitions_organization_id ON requisitions(organization_id);

-- Vendors active index (checked by validator)
CREATE INDEX IF NOT EXISTS idx_vendors_active ON vendors(active);

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 005_add_validator_indexes completed successfully';
END $$;
