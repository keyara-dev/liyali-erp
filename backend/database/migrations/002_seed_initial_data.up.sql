-- ============================================================================
-- LIYALI GATEWAY - SEED INITIAL DATA
-- Migration: 002_seed_initial_data
-- Description: Seeds database with initial organizations, users, workflows, and master data
-- Date: 2025-01-07
-- ============================================================================

-- ============================================================================
-- INITIAL ORGANIZATIONS
-- ============================================================================

-- Create default organization
INSERT INTO organizations (id, name, slug, description, logo_url, primary_color, active, tier, created_by, created_at, updated_at)
VALUES 
    ('org-default-001', 'Default Organization', 'default-org', 'Default organization for initial setup', NULL, '#0066CC', true, 'starter', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Create demo organization for testing
INSERT INTO organizations (id, name, slug, description, logo_url, primary_color, active, tier, created_by, created_at, updated_at)
VALUES 
    ('org-demo-001', 'Demo Corporation', 'demo-corp', 'Demo organization for testing and development', NULL, '#FF6B35', true, 'enterprise', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- INITIAL USERS
-- ============================================================================

-- Create system admin user (password: admin123)
INSERT INTO users (id, email, name, password, role, active, current_organization_id, is_super_admin, preferences, created_at, updated_at)
VALUES 
    ('user-admin-001', 'admin@liyali.com', 'System Administrator', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', true, 'org-default-001', true, '{"theme": "light", "language": "en"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Create demo users for different roles
INSERT INTO users (id, email, name, password, role, active, current_organization_id, is_super_admin, preferences, created_at, updated_at)
VALUES 
    ('user-requester-001', 'requester@demo.com', 'John Requester', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'requester', true, 'org-demo-001', false, '{"theme": "light", "language": "en"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('user-approver-001', 'approver@demo.com', 'Jane Approver', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'approver', true, 'org-demo-001', false, '{"theme": "light", "language": "en"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('user-finance-001', 'finance@demo.com', 'Bob Finance', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'finance', true, 'org-demo-001', false, '{"theme": "light", "language": "en"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('user-manager-001', 'manager@demo.com', 'Alice Manager', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'department_manager', true, 'org-demo-001', false, '{"theme": "light", "language": "en"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ORGANIZATION SETTINGS
-- ============================================================================

-- Default organization settings
INSERT INTO organization_settings (id, organization_id, require_digital_signatures, default_approval_chain, currency, fiscal_year_start, enable_budget_validation, budget_variance_threshold, created_at, updated_at)
VALUES 
    ('settings-default-001', 'org-default-001', true, 'requester->approver->finance', 'USD', 1, true, 5.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('settings-demo-001', 'org-demo-001', true, 'requester->department_manager->finance->approver', 'USD', 1, true, 10.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ORGANIZATION MEMBERS
-- ============================================================================

-- Add users to organizations
INSERT INTO organization_members (id, organization_id, user_id, role, department, title, active, joined_at, created_at, updated_at)
VALUES 
    ('member-001', 'org-default-001', 'user-admin-001', 'admin', 'IT', 'System Administrator', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-002', 'org-demo-001', 'user-requester-001', 'requester', 'Operations', 'Operations Specialist', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-003', 'org-demo-001', 'user-approver-001', 'approver', 'Management', 'Department Head', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-004', 'org-demo-001', 'user-finance-001', 'finance', 'Finance', 'Finance Officer', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('member-005', 'org-demo-001', 'user-manager-001', 'department_manager', 'Operations', 'Operations Manager', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ORGANIZATION DEPARTMENTS
-- ============================================================================

-- Create departments for demo organization
INSERT INTO organization_departments (id, organization_id, name, code, description, active, created_at, updated_at)
VALUES 
    ('dept-001', 'org-demo-001', 'Operations', 'OPS', 'Operations and logistics department', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-002', 'org-demo-001', 'Finance', 'FIN', 'Finance and accounting department', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-003', 'org-demo-001', 'Human Resources', 'HR', 'Human resources department', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-004', 'org-demo-001', 'Information Technology', 'IT', 'IT and systems department', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dept-005', 'org-demo-001', 'Procurement', 'PROC', 'Procurement and purchasing department', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- MASTER DATA - VENDORS
-- ============================================================================

-- Create sample vendors (organization-scoped for multi-tenant security)
INSERT INTO vendors (id, organization_id, vendor_code, name, email, phone, country, city, bank_account, tax_id, active, created_by, created_at, updated_at)
VALUES 
    ('vendor-001', 'org-demo-001', 'VEND-001', 'Office Supplies Inc.', 'contact@officesupplies.com', '+1-555-0101', 'United States', 'New York', 'ACC-001-123456', 'TAX-001', true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-002', 'org-demo-001', 'VEND-002', 'Tech Solutions Ltd.', 'sales@techsolutions.com', '+1-555-0102', 'United States', 'San Francisco', 'ACC-002-789012', 'TAX-002', true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-003', 'org-demo-001', 'VEND-003', 'Facility Services Corp.', 'info@facilityservices.com', '+1-555-0103', 'United States', 'Chicago', 'ACC-003-345678', 'TAX-003', true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-004', 'org-demo-001', 'VEND-004', 'Catering Solutions', 'orders@catering.com', '+1-555-0104', 'United States', 'Los Angeles', 'ACC-004-901234', 'TAX-004', true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('vendor-005', 'org-demo-001', 'VEND-005', 'Equipment Rental Co.', 'rentals@equipment.com', '+1-555-0105', 'United States', 'Dallas', 'ACC-005-567890', 'TAX-005', true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- MASTER DATA - CATEGORIES
-- ============================================================================

-- Create categories for demo organization
INSERT INTO categories (id, organization_id, name, description, active, created_at, updated_at)
VALUES 
    ('cat-001', 'org-demo-001', 'Office Supplies', 'General office supplies and stationery', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-002', 'org-demo-001', 'IT Equipment', 'Computer hardware and software', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-003', 'org-demo-001', 'Facility Maintenance', 'Building maintenance and repairs', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-004', 'org-demo-001', 'Professional Services', 'Consulting and professional services', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-005', 'org-demo-001', 'Travel & Entertainment', 'Business travel and entertainment expenses', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cat-006', 'org-demo-001', 'Marketing & Advertising', 'Marketing materials and advertising', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- CATEGORY BUDGET CODES
-- ============================================================================

-- Link categories to budget codes
INSERT INTO category_budget_codes (id, category_id, budget_code, active, created_at, updated_at)
VALUES 
    ('cbc-001', 'cat-001', 'BUDGET-OFC-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cbc-002', 'cat-001', 'BUDGET-OFC-002', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cbc-003', 'cat-002', 'BUDGET-IT-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cbc-004', 'cat-002', 'BUDGET-IT-002', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cbc-005', 'cat-003', 'BUDGET-FAC-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cbc-006', 'cat-004', 'BUDGET-SVC-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cbc-007', 'cat-005', 'BUDGET-TRV-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cbc-008', 'cat-006', 'BUDGET-MKT-001', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE BUDGETS
-- ============================================================================

-- Create sample budgets for demo organization
INSERT INTO budgets (id, organization_id, owner_id, budget_code, name, description, department, department_id, status, fiscal_year, total_budget, allocated_amount, remaining_amount, currency, approval_stage, created_by, created_at, updated_at)
VALUES 
    ('budget-001', 'org-demo-001', 'user-manager-001', 'BUDGET-OFC-001', 'Office Supplies 2025', 'Annual budget for office supplies', 'Operations', 'dept-001', 'approved', '2025', 50000.00, 15000.00, 35000.00, 'USD', 0, 'user-manager-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-002', 'org-demo-001', 'user-manager-001', 'BUDGET-IT-001', 'IT Equipment 2025', 'Annual budget for IT equipment and software', 'Information Technology', 'dept-004', 'approved', '2025', 100000.00, 25000.00, 75000.00, 'USD', 0, 'user-manager-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-003', 'org-demo-001', 'user-finance-001', 'BUDGET-FAC-001', 'Facility Maintenance 2025', 'Annual budget for facility maintenance', 'Operations', 'dept-001', 'approved', '2025', 75000.00, 20000.00, 55000.00, 'USD', 0, 'user-finance-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('budget-004', 'org-demo-001', 'user-manager-001', 'BUDGET-MKT-001', 'Marketing 2025', 'Annual budget for marketing and advertising', 'Operations', 'dept-001', 'draft', '2025', 80000.00, 0.00, 80000.00, 'USD', 0, 'user-manager-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE WORKFLOWS
-- ============================================================================

-- Create default workflows for different document types
INSERT INTO workflows (id, organization_id, name, description, document_type, entity_type, version, stages, is_active, is_default, created_by, created_at, updated_at)
VALUES 
    (gen_random_uuid(), 'org-demo-001', 'Standard Requisition Approval', 'Standard 3-stage approval workflow for requisitions', 'requisition', 'requisition', 1, 
     '[
       {"stage": 1, "name": "Department Manager Review", "approver_role": "department_manager", "required": true},
       {"stage": 2, "name": "Finance Review", "approver_role": "finance", "required": true},
       {"stage": 3, "name": "Final Approval", "approver_role": "approver", "required": true}
     ]'::jsonb, 
     true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
     
    (gen_random_uuid(), 'org-demo-001', 'Express Requisition Approval', 'Fast-track 2-stage approval for low-value requisitions', 'requisition', 'requisition', 1,
     '[
       {"stage": 1, "name": "Manager Review", "approver_role": "department_manager", "required": true},
       {"stage": 2, "name": "Final Approval", "approver_role": "approver", "required": true}
     ]'::jsonb,
     true, false, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
     
    (gen_random_uuid(), 'org-demo-001', 'Standard Purchase Order Approval', 'Standard approval workflow for purchase orders', 'purchase_order', 'purchase_order', 1,
     '[
       {"stage": 1, "name": "Finance Review", "approver_role": "finance", "required": true},
       {"stage": 2, "name": "Final Approval", "approver_role": "approver", "required": true}
     ]'::jsonb,
     true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
     
    (gen_random_uuid(), 'org-demo-001', 'Standard Payment Voucher Approval', 'Standard approval workflow for payment vouchers', 'payment_voucher', 'payment_voucher', 1,
     '[
       {"stage": 1, "name": "Finance Review", "approver_role": "finance", "required": true},
       {"stage": 2, "name": "Final Approval", "approver_role": "approver", "required": true}
     ]'::jsonb,
     true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
     
    (gen_random_uuid(), 'org-demo-001', 'Standard Budget Approval', 'Standard approval workflow for budgets', 'budget', 'budget', 1,
     '[
       {"stage": 1, "name": "Finance Review", "approver_role": "finance", "required": true},
       {"stage": 2, "name": "Management Approval", "approver_role": "approver", "required": true}
     ]'::jsonb,
     true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
     
    (gen_random_uuid(), 'org-demo-001', 'Standard GRN Approval', 'Standard approval workflow for goods received notes', 'grn', 'goods_received_note', 1,
     '[
       {"stage": 1, "name": "Quality Check", "approver_role": "department_manager", "required": true},
       {"stage": 2, "name": "Final Approval", "approver_role": "approver", "required": true}
     ]'::jsonb,
     true, true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- ORGANIZATION ROLES
-- ============================================================================

-- Create custom organization roles
INSERT INTO organization_roles (id, organization_id, name, description, is_system_role, permissions, active, created_by, created_at, updated_at)
VALUES 
    (gen_random_uuid(), 'org-demo-001', 'Procurement Manager', 'Manages procurement processes and vendor relationships', false, 
     '["manage_vendors", "approve_requisitions", "create_purchase_orders", "view_reports"]'::jsonb, 
     true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
     
    (gen_random_uuid(), 'org-demo-001', 'Budget Controller', 'Controls budget allocations and spending', false,
     '["manage_budgets", "view_financial_reports", "approve_budget_transfers", "monitor_spending"]'::jsonb,
     true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
     
    (gen_random_uuid(), 'org-demo-001', 'Warehouse Manager', 'Manages warehouse operations and inventory', false,
     '["manage_inventory", "process_grn", "manage_warehouse_locations", "view_inventory_reports"]'::jsonb,
     true, 'user-admin-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- SAMPLE REQUISITIONS
-- ============================================================================

-- Create sample requisitions for demo
INSERT INTO requisitions (id, organization_id, req_number, requester_id, title, description, department, department_id, status, priority, total_amount, currency, category_id, preferred_vendor_id, created_by, created_by_name, created_by_role, cost_center, project_code, created_at, updated_at)
VALUES 
    ('req-001', 'org-demo-001', 'REQ-2025-001', 'user-requester-001', 'Office Supplies Q1 2025', 'Quarterly office supplies including paper, pens, and folders', 'Operations', 'dept-001', 'draft', 'medium', 2500.00, 'USD', 'cat-001', 'vendor-001', 'user-requester-001', 'John Requester', 'requester', 'CC-OPS-001', 'PROJ-2025-001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    ('req-002', 'org-demo-001', 'REQ-2025-002', 'user-requester-001', 'New Laptops for Development Team', 'Purchase 5 new laptops for the development team', 'Information Technology', 'dept-004', 'pending', 'high', 15000.00, 'USD', 'cat-002', 'vendor-002', 'user-requester-001', 'John Requester', 'requester', 'CC-IT-001', 'PROJ-2025-002', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    ('req-003', 'org-demo-001', 'REQ-2025-003', 'user-requester-001', 'Conference Room Renovation', 'Renovation of main conference room including furniture and AV equipment', 'Operations', 'dept-001', 'approved', 'medium', 8500.00, 'USD', 'cat-003', 'vendor-003', 'user-requester-001', 'John Requester', 'requester', 'CC-OPS-002', 'PROJ-2025-003', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- COMPLETION LOG
-- ============================================================================

-- Log successful completion
DO $$
DECLARE
    org_count INTEGER;
    user_count INTEGER;
    vendor_count INTEGER;
    category_count INTEGER;
    budget_count INTEGER;
    workflow_count INTEGER;
    requisition_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO org_count FROM organizations;
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO vendor_count FROM vendors;
    SELECT COUNT(*) INTO category_count FROM categories;
    SELECT COUNT(*) INTO budget_count FROM budgets;
    SELECT COUNT(*) INTO workflow_count FROM workflows;
    SELECT COUNT(*) INTO requisition_count FROM requisitions;
    
    RAISE NOTICE 'Migration 002_seed_initial_data completed successfully';
    RAISE NOTICE 'Seeded % organizations', org_count;
    RAISE NOTICE 'Seeded % users', user_count;
    RAISE NOTICE 'Seeded % vendors', vendor_count;
    RAISE NOTICE 'Seeded % categories', category_count;
    RAISE NOTICE 'Seeded % budgets', budget_count;
    RAISE NOTICE 'Seeded % workflows', workflow_count;
    RAISE NOTICE 'Seeded % sample requisitions', requisition_count;
    RAISE NOTICE 'Database is ready for use with sample data';
END $$;