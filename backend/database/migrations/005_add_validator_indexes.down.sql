-- ============================================================================
-- LIYALI GATEWAY - ROLLBACK VALIDATOR INDEXES
-- Migration: 005_add_validator_indexes (DOWN)
-- ============================================================================

DROP INDEX IF EXISTS idx_organizations_active;
DROP INDEX IF EXISTS idx_requisitions_organization_id;
DROP INDEX IF EXISTS idx_vendors_active;
