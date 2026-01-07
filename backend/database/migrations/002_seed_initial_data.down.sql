-- ============================================================================
-- LIYALI GATEWAY - ROLLBACK SEED DATA
-- Migration: 002_seed_initial_data.down.sql
-- Description: Removes all seeded data from the database
-- Date: 2025-01-07
-- ============================================================================

-- Delete sample requisitions
DELETE FROM requisitions WHERE organization_id IN ('org-default-001', 'org-demo-001');

-- Delete organization roles
DELETE FROM organization_roles WHERE organization_id IN ('org-default-001', 'org-demo-001');

-- Delete workflows
DELETE FROM workflows WHERE organization_id IN ('org-default-001', 'org-demo-001');

-- Delete budgets
DELETE FROM budgets WHERE organization_id IN ('org-default-001', 'org-demo-001');

-- Delete category budget codes
DELETE FROM category_budget_codes WHERE category_id IN (
    SELECT id FROM categories WHERE organization_id IN ('org-default-001', 'org-demo-001')
);

-- Delete categories
DELETE FROM categories WHERE organization_id IN ('org-default-001', 'org-demo-001');

-- Delete vendors (global vendors created during seeding)
DELETE FROM vendors WHERE created_by IN ('user-admin-001');

-- Delete organization departments
DELETE FROM organization_departments WHERE organization_id IN ('org-default-001', 'org-demo-001');

-- Delete organization members
DELETE FROM organization_members WHERE organization_id IN ('org-default-001', 'org-demo-001');

-- Delete organization settings
DELETE FROM organization_settings WHERE organization_id IN ('org-default-001', 'org-demo-001');

-- Delete users (seeded users)
DELETE FROM users WHERE id IN (
    'user-admin-001', 
    'user-requester-001', 
    'user-approver-001', 
    'user-finance-001', 
    'user-manager-001'
);

-- Delete organizations
DELETE FROM organizations WHERE id IN ('org-default-001', 'org-demo-001');

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 002_seed_initial_data rollback completed successfully';
    RAISE NOTICE 'All seeded data has been removed from the database';
    RAISE NOTICE 'Database is now in clean state with empty tables';
END $$;