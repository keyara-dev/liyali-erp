-- ============================================================================
-- LIYALI GATEWAY - CONSOLIDATED SEED DATA ROLLBACK
-- Migration: 002_consolidated_seed_data.down
-- Description: Rollback of consolidated seed data
-- Date: January 11, 2026
-- ============================================================================

-- ============================================================================
-- ROLLBACK SEED DATA IN REVERSE ORDER
-- ============================================================================

-- Remove user roles
DELETE FROM user_roles WHERE id IN (
    'user-role-001', 'user-role-002', 'user-role-003', 'user-role-004', 'user-role-005'
);

-- Remove role permissions
DELETE FROM role_permissions WHERE role_id IN (
    'role-admin', 'role-manager', 'role-finance', 'role-approver', 'role-requester'
);

-- Remove roles
DELETE FROM roles WHERE id IN (
    'role-admin', 'role-manager', 'role-finance', 'role-approver', 'role-requester'
);

-- Remove permissions
DELETE FROM permissions WHERE id LIKE 'perm-%';

-- Remove purchase order items
DELETE FROM purchase_order_items WHERE id LIKE 'po-item-%';

-- Remove purchase orders
DELETE FROM purchase_orders WHERE id IN ('po-001', 'po-002');

-- Remove requisition items
DELETE FROM requisition_items WHERE id LIKE 'item-%';

-- Remove requisitions
DELETE FROM requisitions WHERE id IN ('req-001', 'req-002', 'req-003');

-- Remove budgets
DELETE FROM budgets WHERE id LIKE 'budget-%';

-- Remove workflow stages
DELETE FROM workflow_stages WHERE id LIKE 'stage-%';

-- Remove workflows
DELETE FROM workflows WHERE id LIKE 'workflow-%';

-- Remove vendors
DELETE FROM vendors WHERE id LIKE 'vendor-%';

-- Remove categories
DELETE FROM categories WHERE id LIKE 'cat-%';

-- Remove budget codes
DELETE FROM budget_codes WHERE id LIKE 'budget-%';

-- Remove organization departments
DELETE FROM organization_departments WHERE id LIKE 'dept-%';

-- Remove organization members
DELETE FROM organization_members WHERE id LIKE 'member-%';

-- Remove organization settings
DELETE FROM organization_settings WHERE id IN (
    'settings-default-001', 'settings-demo-001'
);

-- Remove demo users (keep system admin)
DELETE FROM users WHERE id IN (
    'user-requester-001', 'user-approver-001', 'user-finance-001', 'user-manager-001'
);

-- Remove demo organization (keep default)
DELETE FROM organizations WHERE id = 'org-demo-001';

-- Log completion
DO $$ 
BEGIN 
    RAISE NOTICE 'Seed data rollback completed successfully!';
    RAISE NOTICE 'Removed all demo data while preserving system admin and default organization';
END $$;