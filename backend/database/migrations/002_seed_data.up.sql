-- ============================================================================
-- LIYALI GATEWAY - COMPREHENSIVE SEED DATA
-- Migration: 002_seed_data
-- Description: Complete seed data with organizations, users, workflows, and test approval tasks
-- Date: January 13, 2026
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
INSERT INTO organization_departments (id, organization_id, name, code, description, manager_name, active, is_active, created_at, updated_at)
VALUES 
    ('dept-001', 'org-demo-001', 'Information Technology', 'IT', 'IT Department responsible for technology infrastructure', 'Alice Manager', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-002', 'org-demo-001', 'Finance', 'FIN', 'Finance Department handling budgets and payments', 'Bob Finance', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-003', 'org-demo-001', 'Operations', 'OPS', 'Operations Department managing daily operations', 'Jane Approver', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-004', 'org-demo-001', 'Human Resources', 'HR', 'HR Department managing employee relations', 'Alice Manager', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-005', 'org-demo-001', 'Procurement', 'PROC', 'Procurement Department handling vendor relations', 'Jane Approver', true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
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
INSERT INTO requisitions (id, organization_id, document_number, requester_id, title, description, department, total_amount, currency, status, priority, category_id, created_at, updated_at)
VALUES 
    ('req-001', 'org-demo-001', 'REQ-260111-001', 'user-requester-001', 'New Laptop for Development Team', 'Request for high-performance laptop for software development', 'IT', 2500.00, 'USD', 'draft', 'medium', 'cat-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('req-002', 'org-demo-001', 'REQ-260111-002', 'user-requester-001', 'Office Supplies Replenishment', 'Monthly office supplies replenishment', 'Operations', 500.00, 'USD', 'draft', 'low', 'cat-003', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('req-003', 'org-demo-001', 'REQ-260111-003', 'user-requester-001', 'Software License Renewal', 'Annual renewal of development software licenses', 'IT', 5000.00, 'USD', 'submitted', 'high', 'cat-002', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('req-004', 'org-demo-001', 'REQ-260111-004', 'user-requester-001', 'Training Course Registration', 'Professional development course for team members', 'HR', 1200.00, 'USD', 'submitted', 'medium', 'cat-004', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('req-005', 'org-demo-001', 'REQ-260111-005', 'user-requester-001', 'Facility Maintenance Contract', 'Annual facility maintenance and cleaning services', 'Operations', 8000.00, 'USD', 'submitted', 'high', 'cat-006', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE PURCHASE ORDERS
-- ============================================================================

-- Create sample purchase orders
INSERT INTO purchase_orders (id, organization_id, document_number, created_by, vendor_id, title, description, total_amount, currency, status, priority, created_at, updated_at)
VALUES 
    ('po-001', 'org-demo-001', 'PO-260111-001', 'user-requester-001', 'vendor-002', 'Laptop Purchase Order', 'Purchase order for development laptops', 2500.00, 'USD', 'draft', 'medium', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('po-002', 'org-demo-001', 'PO-260111-002', 'user-requester-001', 'vendor-002', 'Software License Purchase', 'Purchase order for software licenses', 5000.00, 'USD', 'submitted', 'high', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- WORKFLOW ASSIGNMENTS FOR SUBMITTED DOCUMENTS
-- ============================================================================

-- Create workflow assignments for submitted requisitions
INSERT INTO workflow_assignments (
    id,
    organization_id,
    entity_id,
    entity_type,
    workflow_id,
    workflow_version,
    current_stage,
    status,
    stage_history,
    assigned_at,
    assigned_by,
    created_at,
    updated_at
) VALUES 
    ('wa-req-260111-003', 'org-demo-001', 'req-003', 'requisition', '550e8400-e29b-41d4-a716-446655440001', 1, 1, 'in_progress', '[]'::jsonb, CURRENT_TIMESTAMP, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('wa-req-260111-004', 'org-demo-001', 'req-004', 'requisition', '550e8400-e29b-41d4-a716-446655440001', 1, 1, 'in_progress', '[]'::jsonb, CURRENT_TIMESTAMP, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('wa-req-260111-005', 'org-demo-001', 'req-005', 'requisition', '550e8400-e29b-41d4-a716-446655440001', 1, 1, 'in_progress', '[]'::jsonb, CURRENT_TIMESTAMP, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('wa-po-260111-002', 'org-demo-001', 'po-002', 'purchase_order', '550e8400-e29b-41d4-a716-446655440002', 1, 1, 'in_progress', '[]'::jsonb, CURRENT_TIMESTAMP, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;
-- ============================================================================
-- WORKFLOW TASKS FOR TESTING
-- ============================================================================

-- Create workflow tasks for submitted documents
INSERT INTO workflow_tasks (
    id,
    organization_id,
    workflow_assignment_id,
    entity_id,
    entity_type,
    stage_number,
    stage_name,
    assignment_type,
    assigned_role,
    assigned_user_id,
    status,
    priority,
    created_at,
    claimed_at,
    claimed_by,
    completed_at,
    due_date,
    version,
    updated_by,
    claim_expiry
) VALUES 
    ('wt-req-260111-003-stage1', 'org-demo-001', 'wa-req-260111-003', 'req-003', 'requisition', 1, 'Manager Approval', 'role', 'approver', NULL, 'pending', 'high', CURRENT_TIMESTAMP, NULL, NULL, NULL, CURRENT_TIMESTAMP + INTERVAL '3 days', 1, NULL, NULL),
    ('wt-req-260111-004-stage1', 'org-demo-001', 'wa-req-260111-004', 'req-004', 'requisition', 1, 'Manager Approval', 'role', 'approver', NULL, 'pending', 'medium', CURRENT_TIMESTAMP, NULL, NULL, NULL, CURRENT_TIMESTAMP + INTERVAL '2 days', 1, NULL, NULL),
    ('wt-req-260111-005-stage1', 'org-demo-001', 'wa-req-260111-005', 'req-005', 'requisition', 1, 'Manager Approval', 'role', 'approver', NULL, 'pending', 'high', CURRENT_TIMESTAMP, NULL, NULL, NULL, CURRENT_TIMESTAMP + INTERVAL '1 day', 1, NULL, NULL),
    ('wt-po-260111-002-stage1', 'org-demo-001', 'wa-po-260111-002', 'po-002', 'purchase_order', 1, 'Manager Approval', 'role', 'approver', NULL, 'pending', 'high', CURRENT_TIMESTAMP, NULL, NULL, NULL, CURRENT_TIMESTAMP + INTERVAL '2 days', 1, NULL, NULL)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- APPROVAL TASKS FOR TESTING DIFFERENT SCENARIOS
-- ============================================================================

-- Create approval tasks for backward compatibility and testing
INSERT INTO approval_tasks (
    id,
    organization_id,
    document_id,
    document_type,
    approver_id,
    assigned_to,
    status,
    stage,
    comments,
    signature,
    approved_by,
    approved_at,
    rejected_by,
    rejected_at,
    rejection_reason,
    document_number,
    approver_name,
    priority,
    due_at,
    task_type,
    title,
    workflow_id,
    workflow_name,
    stage_name,
    importance,
    created_at,
    updated_at
) VALUES 
    -- Pending approval tasks
    ('at-req-260111-003-stage1', 'org-demo-001', 'req-003', 'requisition', 'user-approver-001', 'user-approver-001', 'pending', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'REQ-260111-003', 'Jane Approver', 'high', CURRENT_TIMESTAMP + INTERVAL '3 days', 'approval', 'Software License Renewal - Manager Approval Required', '550e8400-e29b-41d4-a716-446655440001', 'Standard Requisition Approval', 'Manager Approval', 'high', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    ('at-req-260111-004-stage1', 'org-demo-001', 'req-004', 'requisition', 'user-manager-001', 'user-manager-001', 'pending', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'REQ-260111-004', 'Alice Manager', 'medium', CURRENT_TIMESTAMP + INTERVAL '2 days', 'approval', 'Training Course Registration - Manager Approval Required', '550e8400-e29b-41d4-a716-446655440001', 'Standard Requisition Approval', 'Manager Approval', 'medium', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    ('at-req-260111-005-stage1', 'org-demo-001', 'req-005', 'requisition', 'user-approver-001', 'user-approver-001', 'pending', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'REQ-260111-005', 'Jane Approver', 'high', CURRENT_TIMESTAMP + INTERVAL '1 day', 'approval', 'Facility Maintenance Contract - Manager Approval Required', '550e8400-e29b-41d4-a716-446655440001', 'Standard Requisition Approval', 'Manager Approval', 'high', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    ('at-po-260111-002-stage1', 'org-demo-001', 'po-002', 'purchase_order', 'user-approver-001', 'user-approver-001', 'pending', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'PO-260111-002', 'Jane Approver', 'high', CURRENT_TIMESTAMP + INTERVAL '2 days', 'approval', 'Software License Purchase - Manager Approval Required', '550e8400-e29b-41d4-a716-446655440002', 'Standard Purchase Order Approval', 'Manager Approval', 'high', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    -- Overdue task for testing
    ('at-req-overdue-001', 'org-demo-001', 'req-001', 'requisition', 'user-approver-001', 'user-approver-001', 'pending', 1, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'REQ-260111-001', 'Jane Approver', 'medium', CURRENT_TIMESTAMP - INTERVAL '2 days', 'approval', 'New Laptop for Development Team - OVERDUE', '550e8400-e29b-41d4-a716-446655440001', 'Standard Requisition Approval', 'Manager Approval', 'medium', CURRENT_TIMESTAMP - INTERVAL '5 days', CURRENT_TIMESTAMP),
    
    -- Approved task for testing
    ('at-req-approved-001', 'org-demo-001', 'req-002', 'requisition', 'user-approver-001', 'user-approver-001', 'approved', 1, 'Approved for office supplies replenishment', 'Jane Approver - Digital Signature', 'user-approver-001', CURRENT_TIMESTAMP - INTERVAL '1 day', NULL, NULL, NULL, 'REQ-260111-002', 'Jane Approver', 'low', CURRENT_TIMESTAMP + INTERVAL '1 day', 'approval', 'Office Supplies Replenishment - APPROVED', '550e8400-e29b-41d4-a716-446655440001', 'Standard Requisition Approval', 'Manager Approval', 'low', CURRENT_TIMESTAMP - INTERVAL '3 days', CURRENT_TIMESTAMP),
    
    -- Finance approval tasks (second stage)
    ('at-req-finance-001', 'org-demo-001', 'req-003', 'requisition', 'user-finance-001', 'user-finance-001', 'pending', 2, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'REQ-260111-003', 'Bob Finance', 'high', CURRENT_TIMESTAMP + INTERVAL '2 days', 'approval', 'Software License Renewal - Finance Approval Required', '550e8400-e29b-41d4-a716-446655440001', 'Standard Requisition Approval', 'Finance Approval', 'high', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;
-- ============================================================================
-- COMPLETION MESSAGE
-- ============================================================================

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Comprehensive seed data migration completed successfully';
    RAISE NOTICE 'Created:';
    RAISE NOTICE '- 2 Organizations (Demo + Enterprise)';
    RAISE NOTICE '- 6 Users (admin, requester, approver, finance, manager, viewer)';
    RAISE NOTICE '- 5 Departments with manager names';
    RAISE NOTICE '- 6 Categories';
    RAISE NOTICE '- 6 Budget codes';
    RAISE NOTICE '- 6 Vendors';
    RAISE NOTICE '- 4 Default workflows with stages';
    RAISE NOTICE '- 4 Sample budgets';
    RAISE NOTICE '- 5 Sample requisitions (3 submitted, 2 draft)';
    RAISE NOTICE '- 2 Sample purchase orders (1 submitted, 1 draft)';
    RAISE NOTICE '- 4 Workflow assignments for submitted documents';
    RAISE NOTICE '- 4 Workflow tasks for testing';
    RAISE NOTICE '- 7 Approval tasks for comprehensive testing:';
    RAISE NOTICE '  * 4 Pending tasks (different priorities and due dates)';
    RAISE NOTICE '  * 1 Overdue task';
    RAISE NOTICE '  * 1 Approved task';
    RAISE NOTICE '  * 1 Finance approval task (second stage)';
    RAISE NOTICE '';
    RAISE NOTICE 'System is ready for comprehensive testing!';
    RAISE NOTICE 'Login credentials (all users password: password):';
    RAISE NOTICE '  Admin: admin@liyali.com';
    RAISE NOTICE '  Requester: requester@liyali.com';
    RAISE NOTICE '  Approver: approver@liyali.com (has 4 pending tasks)';
    RAISE NOTICE '  Finance: finance@liyali.com (has 1 pending finance task)';
    RAISE NOTICE '  Manager: manager@liyali.com (has 1 pending task)';
    RAISE NOTICE '  Viewer: viewer@liyali.com';
    RAISE NOTICE '';
    RAISE NOTICE 'Test scenarios available:';
    RAISE NOTICE '- Pending approvals with different priorities';
    RAISE NOTICE '- Overdue tasks';
    RAISE NOTICE '- Multi-stage approval workflow';
    RAISE NOTICE '- Different document types (requisitions, purchase orders)';
    RAISE NOTICE '- Various approval statuses';
END $$;