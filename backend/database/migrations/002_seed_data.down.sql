-- ============================================================================
-- LIYALI GATEWAY - SEED DATA ROLLBACK
-- Migration: 002_seed_data (DOWN)
-- Description: Remove all seed data inserted in the seed data migration
-- Date: January 13, 2026
-- ============================================================================

-- Remove seed data in reverse dependency order

-- Remove approval tasks
DELETE FROM approval_tasks WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove workflow tasks
DELETE FROM workflow_tasks WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove workflow assignments
DELETE FROM workflow_assignments WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove sample purchase orders
DELETE FROM purchase_orders WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove sample requisitions
DELETE FROM requisitions WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove sample budgets
DELETE FROM budgets WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove workflow defaults
DELETE FROM workflow_defaults WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove default workflows
DELETE FROM workflows WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove vendors
DELETE FROM vendors WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove budget codes
DELETE FROM category_budget_codes WHERE category_id IN (
    SELECT id FROM categories WHERE organization_id IN ('org-demo-001', 'org-enterprise-001')
);

-- Remove categories
DELETE FROM categories WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove organization members
DELETE FROM organization_members WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove organization departments
DELETE FROM organization_departments WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove organization settings
DELETE FROM organization_settings WHERE organization_id IN ('org-demo-001', 'org-enterprise-001');

-- Remove users
DELETE FROM users WHERE id IN (
    'user-admin-001', 'user-requester-001', 'user-approver-001', 
    'user-finance-001', 'user-manager-001', 'user-viewer-001'
);

-- Remove organizations
DELETE FROM organizations WHERE id IN ('org-demo-001', 'org-enterprise-001');

-- Log completion
DO $
BEGIN
    RAISE NOTICE 'Seed data rollback completed successfully';
    RAISE NOTICE 'All seed data has been removed from the database';
END $;