-- ============================================================================
-- LIYALI GATEWAY - ADD MANAGER NAME TO DEPARTMENTS
-- Migration: 002_add_manager_name_to_departments
-- Description: Add manager_name column to organization_departments table
-- Date: January 13, 2026
-- ============================================================================

-- Add manager_name column to organization_departments table
ALTER TABLE organization_departments 
ADD COLUMN manager_name VARCHAR(255);

-- Add index on manager_name for better query performance
CREATE INDEX IF NOT EXISTS idx_org_departments_manager_name 
ON organization_departments(manager_name);