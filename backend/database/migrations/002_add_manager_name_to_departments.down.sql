-- ============================================================================
-- LIYALI GATEWAY - ROLLBACK MANAGER NAME FROM DEPARTMENTS
-- Migration: 002_add_manager_name_to_departments.down
-- Description: Remove manager_name column from organization_departments table
-- Date: January 13, 2026
-- ============================================================================

-- Drop index first
DROP INDEX IF EXISTS idx_org_departments_manager_name;

-- Remove manager_name column from organization_departments table
ALTER TABLE organization_departments 
DROP COLUMN IF EXISTS manager_name;