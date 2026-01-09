-- ============================================================================
-- LIYALI GATEWAY - CONSOLIDATED COMPLETE DATABASE SCHEMA
-- Migration: 001_create_complete_schema_consolidated
-- Description: Creates complete database schema with all fields and enhancements
-- Version: Consolidated from migrations 001, 002, 003, 007
-- Date: 2025-01-07
-- ============================================================================

-- ============================================================================
-- CORE TABLES (No dependencies)
-- ============================================================================

-- Users table (must be first - referenced by many other tables)
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'requester',
    active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    current_organization_id VARCHAR(255),
    is_super_admin BOOLEAN DEFAULT false,
    preferences JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Organizations table (referenced by many other tables)
CREATE TABLE IF NOT EXISTS organizations (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    logo_url VARCHAR(500),
    primary_color VARCHAR(7) DEFAULT '#0066CC',
    active BOOLEAN DEFAULT true,
    tier VARCHAR(20) DEFAULT 'free',
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_organizations_creator FOREIGN KEY (created_by) REFERENCES users(id)
);

-- ============================================================================
-- ORGANIZATION RELATED TABLES
-- ============================================================================

-- Organization Settings
CREATE TABLE IF NOT EXISTS organization_settings (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) UNIQUE NOT NULL,
    require_digital_signatures BOOLEAN DEFAULT true,
    default_approval_chain TEXT,
    currency VARCHAR(3) DEFAULT 'USD',
    fiscal_year_start INTEGER DEFAULT 1,
    enable_budget_validation BOOLEAN DEFAULT true,
    budget_variance_threshold DECIMAL(5,2) DEFAULT 5.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_org_settings_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Organization Members
CREATE TABLE IF NOT EXISTS organization_members (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    department VARCHAR(100),
    title VARCHAR(100),
    active BOOLEAN DEFAULT true,
    invited_at TIMESTAMP,
    joined_at TIMESTAMP,
    invited_by VARCHAR(255),
    custom_permissions JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_org_members_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_org_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_org_members_invited_by FOREIGN KEY (invited_by) REFERENCES users(id),
    CONSTRAINT uk_org_user UNIQUE (organization_id, user_id)
);

-- Organization Departments
CREATE TABLE IF NOT EXISTS organization_departments (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    description TEXT,
    parent_id VARCHAR(255),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_org_departments_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_org_departments_parent FOREIGN KEY (parent_id) REFERENCES organization_departments(id)
);

-- ============================================================================
-- ENHANCED AUTHENTICATION TABLES
-- ============================================================================

-- Sessions table for refresh token management
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(500) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Password reset tokens
CREATE TABLE IF NOT EXISTS password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_password_resets_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Email verification tokens
CREATE TABLE IF NOT EXISTS email_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_email_verifications_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Login attempts tracking for security
CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    success BOOLEAN NOT NULL DEFAULT false,
    failure_reason VARCHAR(255),
    attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_login_attempts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Account lockouts for security
CREATE TABLE IF NOT EXISTS account_lockouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    reason VARCHAR(255) NOT NULL,
    locked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    unlocks_at TIMESTAMP NOT NULL,
    active BOOLEAN DEFAULT true,
    
    CONSTRAINT fk_account_lockouts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Organization roles (custom roles within organizations)
CREATE TABLE IF NOT EXISTS organization_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT false,
    permissions JSONB DEFAULT '[]'::jsonb,
    active BOOLEAN DEFAULT true,
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_org_roles_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_org_roles_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT uk_org_role_name UNIQUE (organization_id, name)
);

-- User role assignments (many-to-many with organizations)
CREATE TABLE IF NOT EXISTS user_organization_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL,
    assigned_by VARCHAR(255),
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT true,
    
    CONSTRAINT fk_user_org_roles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_org_roles_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_org_roles_role_id FOREIGN KEY (role_id) REFERENCES organization_roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_org_roles_assigned_by FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT uk_user_org_role UNIQUE (user_id, organization_id, role_id)
);

-- ============================================================================
-- ENHANCED WORKFLOW SYSTEM TABLES
-- ============================================================================

-- Workflow definitions (organization-specific) - Enhanced from 007_fix_workflow_tables.sql
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    document_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL, -- Added from 007: Type of entity this workflow applies to
    version INTEGER DEFAULT 1, -- Added from 007: Workflow version
    stages JSONB NOT NULL DEFAULT '[]'::jsonb,
    conditions JSONB, -- Added from 007: JSON conditions for when this workflow should be applied
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP, -- Added from 007: Soft delete support
    
    CONSTRAINT fk_workflows_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflows_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT uk_org_workflow_name UNIQUE (organization_id, name)
);

-- Workflow assignments - New from 007_fix_workflow_tables.sql
CREATE TABLE IF NOT EXISTS workflow_assignments (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    workflow_id UUID NOT NULL,
    workflow_version INTEGER NOT NULL,
    current_stage INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'in_progress',
    stage_history JSONB, -- JSON array of completed stage executions
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by VARCHAR(255) NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_workflow_assignments_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_assignments_workflow_id FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_assignments_assigned_by FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Workflow tasks - New from 007_fix_workflow_tables.sql
CREATE TABLE IF NOT EXISTS workflow_tasks (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    workflow_assignment_id VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    stage_number INTEGER NOT NULL,
    stage_name VARCHAR(255) NOT NULL,
    assignment_type VARCHAR(50) DEFAULT 'role', -- How the task is assigned: role or specific_user
    assigned_role VARCHAR(100),
    assigned_user_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    priority VARCHAR(50) DEFAULT 'medium',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    claimed_at TIMESTAMP,
    claimed_by VARCHAR(255),
    completed_at TIMESTAMP,
    due_date TIMESTAMP,
    
    CONSTRAINT fk_workflow_tasks_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_tasks_assignment_id FOREIGN KEY (workflow_assignment_id) REFERENCES workflow_assignments(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_tasks_assigned_user_id FOREIGN KEY (assigned_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_workflow_tasks_claimed_by FOREIGN KEY (claimed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Workflow defaults - New from 007_fix_workflow_tables.sql
CREATE TABLE IF NOT EXISTS workflow_defaults (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    default_workflow_id UUID NOT NULL,
    default_workflow_version INTEGER NOT NULL,
    set_by VARCHAR(255) NOT NULL,
    set_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_workflow_defaults_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_defaults_workflow_id FOREIGN KEY (default_workflow_id) REFERENCES workflows(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_defaults_set_by FOREIGN KEY (set_by) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uk_workflow_defaults_unique_org_entity UNIQUE (organization_id, entity_type)
);

-- Enhanced approval tasks with workflow support
CREATE TABLE IF NOT EXISTS approval_tasks_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    document_id VARCHAR(255) NOT NULL,
    document_type VARCHAR(100) NOT NULL,
    workflow_id UUID,
    assigned_to VARCHAR(255) NOT NULL,
    assigned_by VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    current_stage INTEGER NOT NULL DEFAULT 1,
    total_stages INTEGER NOT NULL DEFAULT 1,
    priority VARCHAR(20) DEFAULT 'MEDIUM',
    due_date TIMESTAMP,
    notes TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_approval_tasks_enh_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_approval_tasks_enh_workflow_id FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE SET NULL,
    CONSTRAINT fk_approval_tasks_enh_assigned_to FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_approval_tasks_enh_assigned_by FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_approval_status CHECK (status IN ('PENDING', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'REASSIGNED')),
    CONSTRAINT chk_approval_priority CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT'))
);

-- Approval history for audit trail
CREATE TABLE IF NOT EXISTS approval_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    stage INTEGER NOT NULL,
    comment TEXT,
    signature TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_approval_history_task_id FOREIGN KEY (task_id) REFERENCES approval_tasks_enhanced(id) ON DELETE CASCADE,
    CONSTRAINT fk_approval_history_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_approval_action CHECK (action IN ('APPROVED', 'REJECTED', 'REASSIGNED', 'COMMENTED', 'VIEWED'))
);

-- Enhanced notifications system
CREATE TABLE IF NOT EXISTS notifications_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    related_id UUID,
    related_type VARCHAR(50),
    is_read BOOLEAN DEFAULT false,
    sent_via_email BOOLEAN DEFAULT false,
    email_sent_at TIMESTAMP,
    priority VARCHAR(20) DEFAULT 'MEDIUM',
    metadata JSONB DEFAULT '{}'::jsonb,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_notifications_enh_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_notifications_enh_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_notification_type CHECK (type IN ('TASK_ASSIGNED', 'TASK_APPROVED', 'TASK_REJECTED', 'TASK_REASSIGNED', 'TASK_COMMENTED', 'DOCUMENT_SUBMITTED', 'SYSTEM_ALERT')),
    CONSTRAINT chk_notification_priority CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT'))
);

-- ============================================================================
-- MASTER DATA TABLES
-- ============================================================================

-- Organization-Scoped Vendors (multi-tenant security)
CREATE TABLE IF NOT EXISTS vendors (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    vendor_code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    country VARCHAR(100),
    city VARCHAR(100),
    bank_account VARCHAR(100),
    tax_id VARCHAR(100),
    active BOOLEAN DEFAULT true,
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_vendors_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_vendors_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT uk_org_vendor_code UNIQUE (organization_id, vendor_code)
);

-- Categories for requisition categorization (organization-specific)
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_categories_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT uk_org_category_name UNIQUE (organization_id, name)
);

-- Category Budget Codes (one-to-many relationship)
CREATE TABLE IF NOT EXISTS category_budget_codes (
    id VARCHAR(255) PRIMARY KEY,
    category_id VARCHAR(255) NOT NULL,
    budget_code VARCHAR(100) NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_category_budget_codes_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

-- ============================================================================
-- BUSINESS DOCUMENT TABLES (Enhanced with all fields from migration 002)
-- ============================================================================

-- Requisitions - Enhanced with all fields from 002_add_missing_fields.up.sql
CREATE TABLE IF NOT EXISTS requisitions (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    req_number VARCHAR(100) UNIQUE NOT NULL,
    requester_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    department VARCHAR(100),
    department_id VARCHAR(255), -- Added from 002
    status VARCHAR(50) DEFAULT 'draft',
    priority VARCHAR(20) DEFAULT 'medium',
    items JSONB,
    total_amount DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    category_id VARCHAR(255),
    preferred_vendor_id VARCHAR(255),
    is_estimate BOOLEAN DEFAULT false,
    -- Additional fields from 002_add_missing_fields.up.sql
    required_by_date TIMESTAMP, -- Added from 002
    cost_center VARCHAR(255), -- Added from 002
    project_code VARCHAR(255), -- Added from 002
    created_by VARCHAR(255), -- Added from 002
    created_by_name VARCHAR(255), -- Added from 002
    created_by_role VARCHAR(255), -- Added from 002
    metadata JSONB, -- Added from 002
    -- Automation fields
    automation_used BOOLEAN DEFAULT FALSE, -- Whether automation was used in processing
    auto_created_po JSONB, -- Auto-created purchase order details
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_requisitions_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_requisitions_requester FOREIGN KEY (requester_id) REFERENCES users(id),
    CONSTRAINT fk_requisitions_category FOREIGN KEY (category_id) REFERENCES categories(id),
    CONSTRAINT fk_requisitions_vendor FOREIGN KEY (preferred_vendor_id) REFERENCES vendors(id),
    CONSTRAINT fk_requisitions_created_by FOREIGN KEY (created_by) REFERENCES users(id) -- Added from 002
);

-- Budgets - Enhanced with all fields from 002_add_missing_fields.up.sql
CREATE TABLE IF NOT EXISTS budgets (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    budget_code VARCHAR(100) NOT NULL,
    department VARCHAR(100),
    department_id VARCHAR(255), -- Added from 002
    status VARCHAR(50) DEFAULT 'draft',
    fiscal_year VARCHAR(10),
    total_budget DECIMAL(15,2),
    allocated_amount DECIMAL(15,2) DEFAULT 0,
    remaining_amount DECIMAL(15,2),
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    -- Additional fields from 002_add_missing_fields.up.sql
    name VARCHAR(255), -- Added from 002
    description TEXT, -- Added from 002
    currency VARCHAR(3) DEFAULT 'USD', -- Added from 002
    created_by VARCHAR(255), -- Added from 002
    items JSONB, -- Added from 002
    action_history JSONB, -- Added from 002
    metadata JSONB, -- Added from 002
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_budgets_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_budgets_owner FOREIGN KEY (owner_id) REFERENCES users(id),
    CONSTRAINT fk_budgets_created_by FOREIGN KEY (created_by) REFERENCES users(id) -- Added from 002
);

-- Purchase Orders - Enhanced with all fields from 002_add_missing_fields.up.sql
CREATE TABLE IF NOT EXISTS purchase_orders (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    po_number VARCHAR(100) UNIQUE NOT NULL,
    vendor_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    items JSONB,
    total_amount DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    delivery_date TIMESTAMP,
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    linked_requisition VARCHAR(255),
    -- Additional fields from 002_add_missing_fields.up.sql
    description TEXT, -- Added from 002
    department VARCHAR(255), -- Added from 002
    department_id VARCHAR(255), -- Added from 002
    gl_code VARCHAR(255), -- Added from 002
    title VARCHAR(255), -- Added from 002
    priority VARCHAR(50) DEFAULT 'medium', -- Added from 002
    subtotal DECIMAL(15,2), -- Added from 002
    tax DECIMAL(15,2), -- Added from 002
    total DECIMAL(15,2), -- Added from 002
    budget_code VARCHAR(255), -- Added from 002
    cost_center VARCHAR(255), -- Added from 002
    project_code VARCHAR(255), -- Added from 002
    required_by_date TIMESTAMP, -- Added from 002
    source_requisition_number VARCHAR(255), -- Added from 002
    source_requisition_id VARCHAR(255), -- Added from 002
    created_by VARCHAR(255), -- Added from 002
    owner_id VARCHAR(255), -- Added from 002
    action_history JSONB, -- Added from 002
    metadata JSONB, -- Added from 002
    -- Automation fields
    automation_used BOOLEAN DEFAULT FALSE, -- Whether automation was used in processing
    auto_created_grn JSONB, -- Auto-created GRN details
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_purchase_orders_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_purchase_orders_vendor FOREIGN KEY (vendor_id) REFERENCES vendors(id),
    CONSTRAINT fk_purchase_orders_created_by FOREIGN KEY (created_by) REFERENCES users(id), -- Added from 002
    CONSTRAINT fk_purchase_orders_source_requisition_id FOREIGN KEY (source_requisition_id) REFERENCES requisitions(id) -- Added from 002
);

-- Payment Vouchers - Enhanced with all fields from 002_add_missing_fields.up.sql
CREATE TABLE IF NOT EXISTS payment_vouchers (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    voucher_number VARCHAR(100) UNIQUE NOT NULL,
    vendor_id VARCHAR(255) NOT NULL,
    invoice_number VARCHAR(100),
    status VARCHAR(50) DEFAULT 'draft',
    amount DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50), -- bank_transfer, cash
    gl_code VARCHAR(100),
    description TEXT,
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    linked_po VARCHAR(255),
    -- Additional fields from 002_add_missing_fields.up.sql
    title VARCHAR(255), -- Added from 002
    department VARCHAR(255), -- Added from 002
    department_id VARCHAR(255), -- Added from 002
    priority VARCHAR(50) DEFAULT 'medium', -- Added from 002
    requested_by_name VARCHAR(255), -- Added from 002
    requested_date TIMESTAMP, -- Added from 002
    submitted_at TIMESTAMP, -- Added from 002
    approved_at TIMESTAMP, -- Added from 002
    paid_date TIMESTAMP, -- Added from 002
    payment_due_date TIMESTAMP, -- Added from 002
    budget_code VARCHAR(255), -- Added from 002
    cost_center VARCHAR(255), -- Added from 002
    project_code VARCHAR(255), -- Added from 002
    tax_amount DECIMAL(15,2), -- Added from 002
    withholding_tax_amount DECIMAL(15,2), -- Added from 002
    paid_amount DECIMAL(15,2), -- Added from 002
    source_purchase_order_number VARCHAR(255), -- Added from 002
    source_requisition_number VARCHAR(255), -- Added from 002
    bank_details JSONB, -- Added from 002
    items JSONB, -- Added from 002
    created_by VARCHAR(255), -- Added from 002
    owner_id VARCHAR(255), -- Added from 002
    action_history JSONB, -- Added from 002
    metadata JSONB, -- Added from 002
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_payment_vouchers_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_payment_vouchers_vendor FOREIGN KEY (vendor_id) REFERENCES vendors(id),
    CONSTRAINT fk_payment_vouchers_created_by FOREIGN KEY (created_by) REFERENCES users(id) -- Added from 002
);

-- Goods Received Notes - Enhanced with all fields from 002_add_missing_fields.up.sql
CREATE TABLE IF NOT EXISTS goods_received_notes (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    grn_number VARCHAR(100) UNIQUE NOT NULL,
    po_number VARCHAR(255),
    status VARCHAR(50) DEFAULT 'draft', -- draft, pending, approved, rejected, paid, completed, cancelled
    received_date TIMESTAMP,
    received_by VARCHAR(255),
    items JSONB,
    quality_issues JSONB,
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    -- Additional fields from 002_add_missing_fields.up.sql
    created_by VARCHAR(255), -- Added from 002
    owner_id VARCHAR(255), -- Added from 002
    warehouse_location VARCHAR(255), -- Added from 002
    notes TEXT, -- Added from 002
    stage_name VARCHAR(255), -- Added from 002
    approved_by VARCHAR(255), -- Added from 002
    automation_used BOOLEAN DEFAULT FALSE, -- Added from 002
    auto_created_pv JSONB, -- Added from 002
    action_history JSONB, -- Added from 002
    metadata JSONB, -- Added from 002
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_grns_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_grns_received_by FOREIGN KEY (received_by) REFERENCES users(id),
    CONSTRAINT fk_grns_po_number FOREIGN KEY (po_number) REFERENCES purchase_orders(po_number),
    CONSTRAINT fk_grn_created_by FOREIGN KEY (created_by) REFERENCES users(id) -- Added from 002
);

-- ============================================================================
-- LEGACY COMPATIBILITY TABLES (Enhanced with fields from migration 002)
-- ============================================================================

-- Legacy Approval Tasks (for backward compatibility) - Enhanced with fields from 002
CREATE TABLE IF NOT EXISTS approval_tasks (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    document_id VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    approver_id VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    stage INTEGER DEFAULT 1,
    comments TEXT,
    signature TEXT,
    approved_by VARCHAR(255),
    approved_at TIMESTAMP,
    rejected_by VARCHAR(255),
    rejected_at TIMESTAMP,
    rejection_reason TEXT,
    -- Additional fields from 002_add_missing_fields.up.sql
    priority VARCHAR(50) DEFAULT 'medium', -- Added from 002
    due_at TIMESTAMP, -- Added from 002
    task_type VARCHAR(100), -- Added from 002
    title VARCHAR(255), -- Added from 002
    workflow_id VARCHAR(255), -- Added from 002
    workflow_name VARCHAR(255), -- Added from 002
    stage_name VARCHAR(255), -- Added from 002
    importance VARCHAR(50) DEFAULT 'medium', -- Added from 002
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_approval_tasks_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_approval_tasks_approver FOREIGN KEY (approver_id) REFERENCES users(id),
    CONSTRAINT fk_approval_tasks_assigned_to FOREIGN KEY (assigned_to) REFERENCES users(id),
    CONSTRAINT fk_approval_tasks_approved_by FOREIGN KEY (approved_by) REFERENCES users(id),
    CONSTRAINT fk_approval_tasks_rejected_by FOREIGN KEY (rejected_by) REFERENCES users(id)
);

-- Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(255) PRIMARY KEY,
    document_id VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    changes JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Legacy Notifications (for backward compatibility)
CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    recipient_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    document_id VARCHAR(255),
    document_type VARCHAR(50),
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    sent BOOLEAN DEFAULT false,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_notifications_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_notifications_recipient FOREIGN KEY (recipient_id) REFERENCES users(id)
);

-- ============================================================================
-- COMPREHENSIVE INDEXES FOR PERFORMANCE
-- ============================================================================

-- Users indexes
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_current_org_id ON users(current_organization_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);

-- Organizations indexes
CREATE INDEX IF NOT EXISTS idx_organizations_active ON organizations(active);
CREATE INDEX IF NOT EXISTS idx_organizations_tier ON organizations(tier);
CREATE INDEX IF NOT EXISTS idx_organizations_created_by ON organizations(created_by);

-- Organization members indexes
CREATE INDEX IF NOT EXISTS idx_org_members_organization_id ON organization_members(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_role ON organization_members(role);
CREATE INDEX IF NOT EXISTS idx_org_members_active ON organization_members(active);

-- Sessions indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Password resets indexes
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);

-- Email verifications indexes
CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token);
CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);

-- Login attempts indexes
CREATE INDEX IF NOT EXISTS idx_login_attempts_email ON login_attempts(email);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_address ON login_attempts(ip_address);
CREATE INDEX IF NOT EXISTS idx_login_attempts_attempted_at ON login_attempts(attempted_at);

-- Account lockouts indexes
CREATE INDEX IF NOT EXISTS idx_account_lockouts_user_id ON account_lockouts(user_id);
CREATE INDEX IF NOT EXISTS idx_account_lockouts_active ON account_lockouts(active);
CREATE INDEX IF NOT EXISTS idx_account_lockouts_unlocks_at ON account_lockouts(unlocks_at);

-- Organization roles indexes
CREATE INDEX IF NOT EXISTS idx_org_roles_organization_id ON organization_roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_roles_active ON organization_roles(active);

-- User organization roles indexes
CREATE INDEX IF NOT EXISTS idx_user_org_roles_user_id ON user_organization_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_org_roles_organization_id ON user_organization_roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_org_roles_active ON user_organization_roles(active);

-- Workflows indexes (Enhanced from 007)
CREATE INDEX IF NOT EXISTS idx_workflows_organization_id ON workflows(organization_id);
CREATE INDEX IF NOT EXISTS idx_workflows_document_type ON workflows(document_type);
CREATE INDEX IF NOT EXISTS idx_workflows_entity_type ON workflows(entity_type); -- Added from 007
CREATE INDEX IF NOT EXISTS idx_workflows_active ON workflows(is_active);
CREATE INDEX IF NOT EXISTS idx_workflows_deleted_at ON workflows(deleted_at); -- Added from 007

-- Workflow assignments indexes (New from 007)
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_organization_id ON workflow_assignments(organization_id);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_entity_id ON workflow_assignments(entity_id);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_entity_type ON workflow_assignments(entity_type);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_workflow_id ON workflow_assignments(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_status ON workflow_assignments(status);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_assigned_by ON workflow_assignments(assigned_by);

-- Workflow tasks indexes (New from 007)
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_organization_id ON workflow_tasks(organization_id);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_assignment_id ON workflow_tasks(workflow_assignment_id);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_entity_id ON workflow_tasks(entity_id);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_status ON workflow_tasks(status);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_assigned_role ON workflow_tasks(assigned_role);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_assigned_user_id ON workflow_tasks(assigned_user_id);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_claimed_by ON workflow_tasks(claimed_by);

-- Workflow defaults indexes (New from 007)
CREATE INDEX IF NOT EXISTS idx_workflow_defaults_organization_id ON workflow_defaults(organization_id);
CREATE INDEX IF NOT EXISTS idx_workflow_defaults_entity_type ON workflow_defaults(entity_type);
CREATE INDEX IF NOT EXISTS idx_workflow_defaults_workflow_id ON workflow_defaults(default_workflow_id);

-- Composite indexes for better query performance (New from 007)
CREATE INDEX IF NOT EXISTS idx_workflows_org_entity_active ON workflows(organization_id, entity_type, is_active);
CREATE INDEX IF NOT EXISTS idx_workflows_org_entity_default ON workflows(organization_id, entity_type, is_default);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_org_entity_status ON workflow_assignments(organization_id, entity_id, entity_type, status);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_org_role_status ON workflow_tasks(organization_id, assigned_role, status);

-- Approval tasks enhanced indexes
CREATE INDEX IF NOT EXISTS idx_approval_tasks_enh_organization_id ON approval_tasks_enhanced(organization_id);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_enh_assigned_to ON approval_tasks_enhanced(assigned_to);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_enh_status ON approval_tasks_enhanced(status);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_enh_document_id ON approval_tasks_enhanced(document_id);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_enh_due_date ON approval_tasks_enhanced(due_date);

-- Approval history indexes
CREATE INDEX IF NOT EXISTS idx_approval_history_task_id ON approval_history(task_id);
CREATE INDEX IF NOT EXISTS idx_approval_history_user_id ON approval_history(user_id);
CREATE INDEX IF NOT EXISTS idx_approval_history_action ON approval_history(action);

-- Notifications enhanced indexes
CREATE INDEX IF NOT EXISTS idx_notifications_enh_organization_id ON notifications_enhanced(organization_id);
CREATE INDEX IF NOT EXISTS idx_notifications_enh_user_id ON notifications_enhanced(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_enh_is_read ON notifications_enhanced(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_enh_type ON notifications_enhanced(type);

-- Business tables indexes (Enhanced from 002)
CREATE INDEX IF NOT EXISTS idx_requisitions_organization_id ON requisitions(organization_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_requester_id ON requisitions(requester_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_status ON requisitions(status);
CREATE INDEX IF NOT EXISTS idx_requisitions_department ON requisitions(department);
-- Additional indexes from 002_add_missing_fields.up.sql
CREATE INDEX IF NOT EXISTS idx_requisitions_department_id ON requisitions(department_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_created_by ON requisitions(created_by);
CREATE INDEX IF NOT EXISTS idx_requisitions_cost_center ON requisitions(cost_center);

CREATE INDEX IF NOT EXISTS idx_budgets_organization_id ON budgets(organization_id);
CREATE INDEX IF NOT EXISTS idx_budgets_owner_id ON budgets(owner_id);
CREATE INDEX IF NOT EXISTS idx_budgets_status ON budgets(status);
CREATE INDEX IF NOT EXISTS idx_budgets_budget_code ON budgets(budget_code);
-- Additional indexes from 002_add_missing_fields.up.sql
CREATE INDEX IF NOT EXISTS idx_budgets_department_id ON budgets(department_id);
CREATE INDEX IF NOT EXISTS idx_budgets_created_by ON budgets(created_by);

CREATE INDEX IF NOT EXISTS idx_purchase_orders_organization_id ON purchase_orders(organization_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_vendor_id ON purchase_orders(vendor_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_status ON purchase_orders(status);
-- Additional indexes from 002_add_missing_fields.up.sql
CREATE INDEX IF NOT EXISTS idx_purchase_orders_department_id ON purchase_orders(department_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_created_by ON purchase_orders(created_by);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_budget_code ON purchase_orders(budget_code);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_cost_center ON purchase_orders(cost_center);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_source_requisition_id ON purchase_orders(source_requisition_id);

CREATE INDEX IF NOT EXISTS idx_payment_vouchers_organization_id ON payment_vouchers(organization_id);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_vendor_id ON payment_vouchers(vendor_id);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_status ON payment_vouchers(status);
-- Additional indexes from 002_add_missing_fields.up.sql
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_department_id ON payment_vouchers(department_id);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_created_by ON payment_vouchers(created_by);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_budget_code ON payment_vouchers(budget_code);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_cost_center ON payment_vouchers(cost_center);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_payment_due_date ON payment_vouchers(payment_due_date);

CREATE INDEX IF NOT EXISTS idx_grns_organization_id ON goods_received_notes(organization_id);
CREATE INDEX IF NOT EXISTS idx_grns_status ON goods_received_notes(status);
-- Additional indexes from 002_add_missing_fields.up.sql
CREATE INDEX IF NOT EXISTS idx_grn_created_by ON goods_received_notes(created_by);
CREATE INDEX IF NOT EXISTS idx_grn_warehouse_location ON goods_received_notes(warehouse_location);

CREATE INDEX IF NOT EXISTS idx_categories_organization_id ON categories(organization_id);
CREATE INDEX IF NOT EXISTS idx_categories_active ON categories(active);

CREATE INDEX IF NOT EXISTS idx_vendors_organization_id ON vendors(organization_id);
CREATE INDEX IF NOT EXISTS idx_vendors_active ON vendors(active);
CREATE INDEX IF NOT EXISTS idx_vendors_created_by ON vendors(created_by);
CREATE INDEX IF NOT EXISTS idx_vendors_active_org ON vendors(active, organization_id);

CREATE INDEX IF NOT EXISTS idx_approval_tasks_organization_id ON approval_tasks(organization_id);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_assigned_to ON approval_tasks(assigned_to);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_status ON approval_tasks(status);
-- Additional indexes from 002_add_missing_fields.up.sql
CREATE INDEX IF NOT EXISTS idx_approval_tasks_priority ON approval_tasks(priority);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_due_at ON approval_tasks(due_at);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_workflow_id ON approval_tasks(workflow_id);

CREATE INDEX IF NOT EXISTS idx_notifications_organization_id ON notifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

-- ============================================================================
-- TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- ============================================================================

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to tables with updated_at columns
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organization_settings_updated_at BEFORE UPDATE ON organization_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organization_members_updated_at BEFORE UPDATE ON organization_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organization_departments_updated_at BEFORE UPDATE ON organization_departments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organization_roles_updated_at BEFORE UPDATE ON organization_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workflows_updated_at BEFORE UPDATE ON workflows
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workflow_assignments_updated_at BEFORE UPDATE ON workflow_assignments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_approval_tasks_enh_updated_at BEFORE UPDATE ON approval_tasks_enhanced
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_requisitions_updated_at BEFORE UPDATE ON requisitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_budgets_updated_at BEFORE UPDATE ON budgets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_purchase_orders_updated_at BEFORE UPDATE ON purchase_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_vouchers_updated_at BEFORE UPDATE ON payment_vouchers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_goods_received_notes_updated_at BEFORE UPDATE ON goods_received_notes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_category_budget_codes_updated_at BEFORE UPDATE ON category_budget_codes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vendors_updated_at BEFORE UPDATE ON vendors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_approval_tasks_updated_at BEFORE UPDATE ON approval_tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

-- Table comments
COMMENT ON TABLE workflows IS 'Enhanced workflow definitions with frontend compatibility';
COMMENT ON TABLE workflow_assignments IS 'Tracks workflow execution for specific entities';
COMMENT ON TABLE workflow_tasks IS 'Individual approval tasks within workflow assignments';
COMMENT ON TABLE workflow_defaults IS 'Default workflow mappings for entity types per organization';

-- Column comments for business fields
COMMENT ON COLUMN workflows.entity_type IS 'Type of entity this workflow applies to (requisition, purchase_order, etc.)';
COMMENT ON COLUMN workflows.conditions IS 'JSON conditions for when this workflow should be applied';
COMMENT ON COLUMN workflows.stages IS 'JSON array of workflow stages with approval requirements';
COMMENT ON COLUMN workflow_assignments.stage_history IS 'JSON array of completed stage executions';
COMMENT ON COLUMN workflow_tasks.assignment_type IS 'How the task is assigned: role or specific_user';

-- Business field comments from 002_add_missing_fields.up.sql
COMMENT ON COLUMN requisitions.department_id IS 'Department ID reference';
COMMENT ON COLUMN requisitions.cost_center IS 'Cost center for accounting';
COMMENT ON COLUMN requisitions.project_code IS 'Project code for tracking';
COMMENT ON COLUMN requisitions.metadata IS 'Generic metadata for extensibility';
COMMENT ON COLUMN requisitions.automation_used IS 'Whether automation was used in processing';
COMMENT ON COLUMN requisitions.auto_created_po IS 'Auto-created purchase order details';

COMMENT ON COLUMN budgets.name IS 'Budget display name';
COMMENT ON COLUMN budgets.items IS 'Budget line items breakdown';
COMMENT ON COLUMN budgets.metadata IS 'Generic metadata for extensibility';

COMMENT ON COLUMN purchase_orders.gl_code IS 'General Ledger code';
COMMENT ON COLUMN purchase_orders.subtotal IS 'Subtotal before tax';
COMMENT ON COLUMN purchase_orders.tax IS 'Tax amount';
COMMENT ON COLUMN purchase_orders.total IS 'Total amount including tax';
COMMENT ON COLUMN purchase_orders.metadata IS 'Generic metadata for extensibility';
COMMENT ON COLUMN purchase_orders.automation_used IS 'Whether automation was used in processing';
COMMENT ON COLUMN purchase_orders.auto_created_grn IS 'Auto-created GRN details';

COMMENT ON COLUMN payment_vouchers.tax_amount IS 'Tax amount for payment';
COMMENT ON COLUMN payment_vouchers.withholding_tax_amount IS 'Withholding tax amount';
COMMENT ON COLUMN payment_vouchers.paid_amount IS 'Actual amount paid';
COMMENT ON COLUMN payment_vouchers.bank_details IS 'Bank details for payment';
COMMENT ON COLUMN payment_vouchers.items IS 'Payment line items breakdown';
COMMENT ON COLUMN payment_vouchers.metadata IS 'Generic metadata for extensibility';

COMMENT ON COLUMN goods_received_notes.automation_used IS 'Whether automation was used in processing';
COMMENT ON COLUMN goods_received_notes.auto_created_pv IS 'Auto-created payment voucher details';
COMMENT ON COLUMN goods_received_notes.metadata IS 'Generic metadata for extensibility';

-- Status and enum comments from 003_add_alignment_fields.up.sql
COMMENT ON COLUMN goods_received_notes.status IS 'Status: draft, pending, approved, rejected, paid, completed, cancelled';
COMMENT ON COLUMN payment_vouchers.payment_method IS 'Payment method: bank_transfer, cash';

-- ============================================================================
-- MIGRATION COMPLETION LOG
-- ============================================================================

-- Log successful completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 001_create_complete_schema_consolidated completed successfully';
    RAISE NOTICE 'Created % tables with all business fields and enhancements', 
        (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public');
    RAISE NOTICE 'Created % indexes for performance optimization', 
        (SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public');
    RAISE NOTICE 'Database schema is ready for production use';
END $$;