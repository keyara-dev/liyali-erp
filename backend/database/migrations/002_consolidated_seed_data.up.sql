-- ============================================================================
-- LIYALI GATEWAY - CONSOLIDATED SEED DATA
-- Migration: 002_consolidated_seed_data
-- Description: Complete seed data with organizations, users, workflows, and master data
-- Date: January 11, 2026
-- ============================================================================

-- ============================================================================
-- INITIAL ORGANIZATIONS
-- ============================================================================

-- Create default organization
INSERT INTO organizations (id, name, slug, description, active, tier, created_at, updated_at)
VALUES 
    ('org-demo-001', 'Liyali Demo Organization', 'liyali-demo', 'Default organization for testing and development', true, 'pro', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('org-enterprise-001', 'Enterprise Corp', 'enterprise-corp', 'Large enterprise organization for testing enterprise features', true, 'enterprise', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- INITIAL USERS
-- ============================================================================

-- Create system admin user (password: password)
INSERT INTO users (id, email, name, password, role, active, current_organization_id, is_super_admin, created_at, updated_at)
VALUES 
    ('user-admin-001', 'admin@liyali.com', 'System Administrator', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', true, 'org-demo-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('user-requester-001', 'requester@liyali.com', 'John Requester', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'requester', true, 'org-demo-001', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('user-approver-001', 'approver@liyali.com', 'Jane Approver', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'approver', true, 'org-demo-001', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('user-finance-001', 'finance@liyali.com', 'Bob Finance', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'finance', true, 'org-demo-001', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('user-manager-001', 'manager@liyali.com', 'Alice Manager', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'approver', true, 'org-demo-001', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('user-viewer-001', 'viewer@liyali.com', 'Charlie Viewer', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'viewer', true, 'org-demo-001', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ORGANIZATION SETTINGS
-- ============================================================================

-- Default organization settings
INSERT INTO organization_settings (id, organization_id, require_digital_signatures, currency, fiscal_year_start, enable_budget_validation, budget_variance_threshold, created_at, updated_at)
VALUES 
    ('settings-001', 'org-demo-001', true, 'USD', 1, true, 5.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ORGANIZATION DEPARTMENTS
-- ============================================================================

-- Create departments for default organization
INSERT INTO organization_departments (id, organization_id, name, code, description, active, is_active, created_at, updated_at)
VALUES 
    ('dept-001', 'org-demo-001', 'Information Technology', 'IT', 'IT Department responsible for technology infrastructure', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-002', 'org-demo-001', 'Finance', 'FIN', 'Finance Department handling budgets and payments', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-003', 'org-demo-001', 'Operations', 'OPS', 'Operations Department managing daily operations', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-004', 'org-demo-001', 'Human Resources', 'HR', 'HR Department managing employee relations', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-005', 'org-demo-001', 'Procurement', 'PROC', 'Procurement Department handling vendor relations', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ORGANIZATION MEMBERS
-- ============================================================================

-- Add users to organizations
INSERT INTO organization_members (id, organization_id, user_id, role, department, department_id, active, joined_at, created_at, updated_at)
VALUES 
    ('member-001', 'org-demo-001', 'user-admin-001', 'admin', 'IT', 'dept-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-002', 'org-demo-001', 'user-requester-001', 'requester', 'Operations', 'dept-003', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-003', 'org-demo-001', 'user-approver-001', 'approver', 'Finance', 'dept-002', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-004', 'org-demo-001', 'user-finance-001', 'finance', 'Finance', 'dept-002', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-005', 'org-demo-001', 'user-manager-001', 'approver', 'Operations', 'dept-003', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-006', 'org-demo-001', 'user-viewer-001', 'viewer', 'IT', 'dept-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- CATEGORIES
-- ============================================================================

-- Create categories for different types of purchases
INSERT INTO categories (id, organization_id, name, description, active, created_at, updated_at)
VALUES 
    ('cat-001', 'org-demo-001', 'Computer Hardware', 'Desktop computers, laptops, servers, and related hardware', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-002', 'org-demo-001', 'Software Licenses', 'Software licenses, subscriptions, and digital tools', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-003', 'org-demo-001', 'Office Supplies', 'General office supplies, stationery, and consumables', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-004', 'org-demo-001', 'Training & Development', 'Employee training, courses, and professional development', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-005', 'org-demo-001', 'Professional Services', 'Consulting, legal, and other professional services', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-006', 'org-demo-001', 'Facilities & Maintenance', 'Building maintenance, utilities, and facility services', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- BUDGET CODES
-- ============================================================================

-- Create budget codes for categories
INSERT INTO category_budget_codes (id, category_id, budget_code, active, created_at, updated_at)
VALUES 
    ('budget-001', 'cat-001', 'IT-EQUIP', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-002', 'cat-002', 'IT-SOFT', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-003', 'cat-003', 'OFFICE-SUP', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-004', 'cat-004', 'HR-TRAIN', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-005', 'cat-005', 'PROF-SERV', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-006', 'cat-006', 'FACILITIES', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- VENDORS
-- ============================================================================

-- Create sample vendors with organization_id
INSERT INTO vendors (id, organization_id, vendor_code, name, email, phone, country, city, active, created_at, updated_at)
VALUES 
    ('vendor-001', 'org-demo-001', 'VEND-001', 'Office Supplies Inc.', 'contact@officesupplies.com', '+1-555-0101', 'United States', 'New York', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-002', 'org-demo-001', 'VEND-002', 'Tech Solutions Ltd.', 'sales@techsolutions.com', '+1-555-0102', 'United States', 'San Francisco', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-003', 'org-demo-001', 'VEND-003', 'Facility Services Corp.', 'info@facilityservices.com', '+1-555-0103', 'United States', 'Chicago', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-004', 'org-demo-001', 'VEND-004', 'Training Solutions', 'training@solutions.com', '+1-555-0104', 'United States', 'Austin', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-005', 'org-demo-001', 'VEND-005', 'Professional Consultants', 'contact@proconsult.com', '+1-555-0105', 'United States', 'Boston', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-006', 'org-demo-001', 'VEND-006', 'Hardware Direct', 'orders@hardwaredirect.com', '+1-555-0106', 'United States', 'Seattle', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- DEFAULT WORKFLOWS
-- ============================================================================

-- Create default workflows for each document type with stages
INSERT INTO workflows (id, organization_id, name, document_type, entity_type, description, stages, is_default, is_active, created_by, created_at, updated_at)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'org-demo-001', 'Standard Requisition Approval', 'requisition', 'requisition', 'Standard approval workflow for requisitions', '[{"stageNumber": 1, "stageName": "Manager Approval", "requiredRole": "approver", "timeoutHours": 24}, {"stageNumber": 2, "stageName": "Finance Approval", "requiredRole": "finance", "timeoutHours": 48}]'::jsonb, true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('550e8400-e29b-41d4-a716-446655440002', 'org-demo-001', 'Standard Purchase Order Approval', 'purchase_order', 'purchase_order', 'Standard approval workflow for purchase orders', '[{"stageNumber": 1, "stageName": "Manager Approval", "requiredRole": "approver", "timeoutHours": 24}, {"stageNumber": 2, "stageName": "Finance Approval", "requiredRole": "finance", "timeoutHours": 48}]'::jsonb, true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('550e8400-e29b-41d4-a716-446655440003', 'org-demo-001', 'Budget Approval Workflow', 'budget', 'budget', 'Standard approval workflow for budgets', '[{"stageNumber": 1, "stageName": "Manager Approval", "requiredRole": "approver", "timeoutHours": 24}, {"stageNumber": 2, "stageName": "Finance Approval", "requiredRole": "finance", "timeoutHours": 48}]'::jsonb, true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('550e8400-e29b-41d4-a716-446655440004', 'org-demo-001', 'Payment Voucher Approval', 'payment_voucher', 'payment_voucher', 'Standard approval workflow for payment vouchers', '[{"stageNumber": 1, "stageName": "Manager Approval", "requiredRole": "approver", "timeoutHours": 24}, {"stageNumber": 2, "stageName": "Finance Approval", "requiredRole": "finance", "timeoutHours": 48}]'::jsonb, true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- WORKFLOW DEFAULTS
-- ============================================================================

-- Set default workflows for each entity type
INSERT INTO workflow_defaults (id, organization_id, entity_type, default_workflow_id, default_workflow_version, set_by, set_at)
VALUES 
    ('default-001', 'org-demo-001', 'requisition', '550e8400-e29b-41d4-a716-446655440001', 1, 'user-admin-001', CURRENT_TIMESTAMP),
    ('default-002', 'org-demo-001', 'purchase_order', '550e8400-e29b-41d4-a716-446655440002', 1, 'user-admin-001', CURRENT_TIMESTAMP),
    ('default-003', 'org-demo-001', 'budget', '550e8400-e29b-41d4-a716-446655440003', 1, 'user-admin-001', CURRENT_TIMESTAMP),
    ('default-004', 'org-demo-001', 'payment_voucher', '550e8400-e29b-41d4-a716-446655440004', 1, 'user-admin-001', CURRENT_TIMESTAMP)
ON CONFLICT (organization_id, entity_type) DO UPDATE SET
    default_workflow_id = EXCLUDED.default_workflow_id,
    default_workflow_version = EXCLUDED.default_workflow_version,
    set_by = EXCLUDED.set_by,
    set_at = EXCLUDED.set_at;

-- ============================================================================
-- SAMPLE BUDGETS
-- ============================================================================

-- Create sample budgets for the current fiscal year
INSERT INTO budgets (id, organization_id, owner_id, budget_code, name, description, total_budget, allocated_amount, remaining_amount, currency, fiscal_year, status, created_by, created_at, updated_at)
VALUES 
    ('budget-it-001', 'org-demo-001', 'user-admin-001', 'IT-EQUIP', 'IT Equipment Budget 2026', 'Annual budget for IT equipment purchases', 50000.00, 0.00, 50000.00, 'USD', '2026', 'active', 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-it-002', 'org-demo-001', 'user-admin-001', 'IT-SOFT', 'IT Software Budget 2026', 'Annual budget for software licenses and subscriptions', 25000.00, 0.00, 25000.00, 'USD', '2026', 'active', 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-ops-001', 'org-demo-001', 'user-admin-001', 'OFFICE-SUP', 'Operations Supplies Budget 2026', 'Annual budget for operational supplies', 15000.00, 0.00, 15000.00, 'USD', '2026', 'active', 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-hr-001', 'org-demo-001', 'user-admin-001', 'HR-TRAIN', 'HR Training Budget 2026', 'Annual budget for employee training and development', 20000.00, 0.00, 20000.00, 'USD', '2026', 'active', 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE REQUISITIONS
-- ============================================================================

-- Create sample requisitions for testing
INSERT INTO requisitions (id, organization_id, req_number, requester_id, title, description, department, total_amount, currency, status, priority, category_id, created_at, updated_at)
VALUES 
    ('req-001', 'org-demo-001', 'REQ-260111-001', 'user-requester-001', 'New Laptop for Development Team', 'Request for high-performance laptop for software development', 'IT', 2500.00, 'USD', 'draft', 'medium', 'cat-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('req-002', 'org-demo-001', 'REQ-260111-002', 'user-requester-001', 'Office Supplies Replenishment', 'Monthly office supplies replenishment', 'Operations', 500.00, 'USD', 'draft', 'low', 'cat-003', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('req-003', 'org-demo-001', 'REQ-260111-003', 'user-requester-001', 'Software License Renewal', 'Annual renewal of development software licenses', 'IT', 5000.00, 'USD', 'submitted', 'high', 'cat-002', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE PURCHASE ORDERS
-- ============================================================================

-- Create sample purchase orders
INSERT INTO purchase_orders (id, organization_id, po_number, created_by, vendor_id, title, description, total_amount, currency, status, priority, created_at, updated_at)
VALUES 
    ('po-001', 'org-demo-001', 'PO-260111-001', 'user-requester-001', 'vendor-002', 'Laptop Purchase Order', 'Purchase order for development laptops', 2500.00, 'USD', 'draft', 'medium', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- COMPLETION MESSAGE
-- ============================================================================

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Consolidated seed data migration completed successfully';
    RAISE NOTICE 'Created:';
    RAISE NOTICE '- 2 Organizations (Demo + Enterprise)';
    RAISE NOTICE '- 6 Users (admin, requester, approver, finance, manager, viewer)';
    RAISE NOTICE '- 5 Departments';
    RAISE NOTICE '- 6 Categories';
    RAISE NOTICE '- 6 Budget codes';
    RAISE NOTICE '- 6 Vendors';
    RAISE NOTICE '- 4 Default workflows with stages';
    RAISE NOTICE '- 4 Sample budgets';
    RAISE NOTICE '- 3 Sample requisitions';
    RAISE NOTICE '- 1 Sample purchase order';
    RAISE NOTICE '';
    RAISE NOTICE 'System is ready for comprehensive testing!';
    RAISE NOTICE 'Login credentials (all users password: password):';
    RAISE NOTICE '  Admin: admin@liyali.com';
    RAISE NOTICE '  Requester: requester@liyali.com';
    RAISE NOTICE '  Approver: approver@liyali.com';
    RAISE NOTICE '  Finance: finance@liyali.com';
    RAISE NOTICE '  Manager: manager@liyali.com';
    RAISE NOTICE '  Viewer: viewer@liyali.com';
END $$;