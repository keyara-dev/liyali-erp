-- ============================================================================
-- LIYALI GATEWAY - SYSTEM INITIALIZATION
-- Migration: 001_init_system
-- Description: Complete database schema with all enhancements and fixes
-- Version: Consolidated from all previous migrations
-- Date: January 13, 2026
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
    tier VARCHAR(20) DEFAULT 'starter',
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_organizations_creator FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT check_organization_tier CHECK (tier IN ('starter', 'pro', 'enterprise'))
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

-- Organization Departments (includes manager_name from migration 002)
CREATE TABLE IF NOT EXISTS organization_departments (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    description TEXT,
    parent_id VARCHAR(255),
    manager_name VARCHAR(255),
    active BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_org_departments_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_org_departments_parent FOREIGN KEY (parent_id) REFERENCES organization_departments(id)
);

-- Organization Members
CREATE TABLE IF NOT EXISTS organization_members (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    department VARCHAR(100),
    department_id VARCHAR(255),
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
    CONSTRAINT fk_org_members_department FOREIGN KEY (department_id) REFERENCES organization_departments(id) ON DELETE SET NULL,
    CONSTRAINT uk_org_user UNIQUE (organization_id, user_id)
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

-- Workflow definitions (organization-specific)
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    document_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    version INTEGER DEFAULT 1,
    stages JSONB NOT NULL DEFAULT '[]'::jsonb,
    conditions JSONB,
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT fk_workflows_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflows_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT uk_org_workflow_name UNIQUE (organization_id, name)
);

-- Workflow assignments
CREATE TABLE IF NOT EXISTS workflow_assignments (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    workflow_id UUID NOT NULL,
    workflow_version INTEGER NOT NULL,
    current_stage INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'in_progress',
    stage_history JSONB,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by VARCHAR(255) NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_workflow_assignments_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_assignments_workflow_id FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_assignments_assigned_by FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Workflow tasks (enhanced with concurrency control)
CREATE TABLE IF NOT EXISTS workflow_tasks (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    workflow_assignment_id VARCHAR(255) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    stage_number INTEGER NOT NULL,
    stage_name VARCHAR(255) NOT NULL,
    assignment_type VARCHAR(50) DEFAULT 'role',
    assigned_role VARCHAR(100),
    assigned_user_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    priority VARCHAR(50) DEFAULT 'medium',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    claimed_at TIMESTAMP,
    claimed_by VARCHAR(255),
    completed_at TIMESTAMP,
    due_date TIMESTAMP,
    version INTEGER DEFAULT 1 NOT NULL,
    updated_by VARCHAR(255),
    claim_expiry TIMESTAMP,
    
    CONSTRAINT fk_workflow_tasks_organization_id FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_tasks_assignment_id FOREIGN KEY (workflow_assignment_id) REFERENCES workflow_assignments(id) ON DELETE CASCADE,
    CONSTRAINT fk_workflow_tasks_assigned_user_id FOREIGN KEY (assigned_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_workflow_tasks_claimed_by FOREIGN KEY (claimed_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_workflow_tasks_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);
-- Workflow defaults
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

-- Stage approval records (for multiple approval support)
CREATE TABLE IF NOT EXISTS stage_approval_records (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    workflow_task_id VARCHAR(255) NOT NULL,
    stage_number INTEGER NOT NULL,
    approver_id VARCHAR(255) NOT NULL,
    approver_name VARCHAR(255) NOT NULL,
    approver_role VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('approved', 'rejected')),
    comments TEXT,
    signature TEXT,
    approved_at TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_stage_approval_organization 
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_stage_approval_workflow_task 
        FOREIGN KEY (workflow_task_id) REFERENCES workflow_tasks(id) ON DELETE CASCADE,
    CONSTRAINT fk_stage_approval_approver 
        FOREIGN KEY (approver_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Task assignment history (for round-robin tracking)
CREATE TABLE IF NOT EXISTS task_assignment_history (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    role VARCHAR(100) NOT NULL,
    assigned_user_id VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_task_assignment_organization 
        FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_task_assignment_user 
        FOREIGN KEY (assigned_user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- ============================================================================
-- MASTER DATA TABLES
-- ============================================================================

-- Organization-Scoped Vendors
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
-- UNIFIED DOCUMENTS TABLE
-- ============================================================================

-- Documents table for unified document management and search
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    document_number VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    amount DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    department VARCHAR(100),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    workflow_id UUID,
    data JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT fk_documents_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_documents_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT fk_documents_updated_by FOREIGN KEY (updated_by) REFERENCES users(id),
    CONSTRAINT fk_documents_workflow FOREIGN KEY (workflow_id) REFERENCES workflows(id)
);
-- ============================================================================
-- BUSINESS DOCUMENT TABLES
-- ============================================================================

-- Requisitions
CREATE TABLE IF NOT EXISTS requisitions (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    document_number VARCHAR(100) NOT NULL,
    requester_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    department VARCHAR(100),
    department_id VARCHAR(255),
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
    required_by_date TIMESTAMP,
    cost_center VARCHAR(255),
    project_code VARCHAR(255),
    created_by VARCHAR(255),
    created_by_name VARCHAR(255),
    created_by_role VARCHAR(255),
    metadata JSONB,
    automation_used BOOLEAN DEFAULT FALSE,
    auto_created_po JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_requisitions_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_requisitions_requester FOREIGN KEY (requester_id) REFERENCES users(id),
    CONSTRAINT fk_requisitions_category FOREIGN KEY (category_id) REFERENCES categories(id),
    CONSTRAINT fk_requisitions_vendor FOREIGN KEY (preferred_vendor_id) REFERENCES vendors(id),
    CONSTRAINT fk_requisitions_created_by FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Budgets
CREATE TABLE IF NOT EXISTS budgets (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    budget_code VARCHAR(100) NOT NULL,
    department VARCHAR(100),
    department_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'draft',
    fiscal_year VARCHAR(10),
    total_budget DECIMAL(15,2),
    allocated_amount DECIMAL(15,2) DEFAULT 0,
    remaining_amount DECIMAL(15,2),
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    name VARCHAR(255),
    description TEXT,
    currency VARCHAR(3) DEFAULT 'USD',
    created_by VARCHAR(255),
    items JSONB,
    action_history JSONB,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_budgets_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_budgets_owner FOREIGN KEY (owner_id) REFERENCES users(id),
    CONSTRAINT fk_budgets_created_by FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Purchase Orders
CREATE TABLE IF NOT EXISTS purchase_orders (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    document_number VARCHAR(100) NOT NULL,
    vendor_id VARCHAR(255), -- NULLABLE
    status VARCHAR(50) DEFAULT 'draft',
    items JSONB,
    total_amount DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    delivery_date TIMESTAMP,
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    linked_requisition VARCHAR(255),
    description TEXT,
    department VARCHAR(255),
    department_id VARCHAR(255),
    gl_code VARCHAR(255),
    title VARCHAR(255),
    priority VARCHAR(50) DEFAULT 'medium',
    subtotal DECIMAL(15,2),
    tax DECIMAL(15,2),
    total DECIMAL(15,2),
    budget_code VARCHAR(255),
    cost_center VARCHAR(255),
    project_code VARCHAR(255),
    required_by_date TIMESTAMP,
    source_requisition_number VARCHAR(255),
    source_requisition_id VARCHAR(255),
    created_by VARCHAR(255),
    owner_id VARCHAR(255),
    action_history JSONB,
    metadata JSONB,
    automation_used BOOLEAN DEFAULT FALSE,
    auto_created_grn JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_purchase_orders_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_purchase_orders_vendor FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE SET NULL,
    CONSTRAINT fk_purchase_orders_created_by FOREIGN KEY (created_by) REFERENCES users(id)
);
-- Payment Vouchers
CREATE TABLE IF NOT EXISTS payment_vouchers (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    document_number VARCHAR(100) NOT NULL,
    vendor_id VARCHAR(255) NOT NULL,
    invoice_number VARCHAR(100),
    status VARCHAR(50) DEFAULT 'draft',
    amount DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'USD',
    payment_method VARCHAR(50),
    gl_code VARCHAR(100),
    description TEXT,
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    linked_po VARCHAR(255),
    title VARCHAR(255),
    department VARCHAR(255),
    department_id VARCHAR(255),
    priority VARCHAR(50) DEFAULT 'medium',
    requested_by_name VARCHAR(255),
    requested_date TIMESTAMP,
    submitted_at TIMESTAMP,
    approved_at TIMESTAMP,
    paid_date TIMESTAMP,
    payment_due_date TIMESTAMP,
    budget_code VARCHAR(255),
    cost_center VARCHAR(255),
    project_code VARCHAR(255),
    tax_amount DECIMAL(15,2),
    withholding_tax_amount DECIMAL(15,2),
    paid_amount DECIMAL(15,2),
    source_purchase_order_number VARCHAR(255),
    source_requisition_number VARCHAR(255),
    bank_details JSONB,
    items JSONB,
    created_by VARCHAR(255),
    owner_id VARCHAR(255),
    action_history JSONB,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_payment_vouchers_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_payment_vouchers_vendor FOREIGN KEY (vendor_id) REFERENCES vendors(id),
    CONSTRAINT fk_payment_vouchers_created_by FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Goods Received Notes
CREATE TABLE IF NOT EXISTS goods_received_notes (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    document_number VARCHAR(100) NOT NULL,
    po_document_number VARCHAR(255),
    status VARCHAR(50) DEFAULT 'draft',
    received_date TIMESTAMP,
    received_by VARCHAR(255),
    items JSONB,
    quality_issues JSONB,
    approval_stage INTEGER DEFAULT 0,
    approval_history JSONB,
    created_by VARCHAR(255),
    owner_id VARCHAR(255),
    warehouse_location VARCHAR(255),
    notes TEXT,
    stage_name VARCHAR(255),
    approved_by VARCHAR(255),
    automation_used BOOLEAN DEFAULT FALSE,
    auto_created_pv JSONB,
    action_history JSONB,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_grns_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_grns_received_by FOREIGN KEY (received_by) REFERENCES users(id),
    CONSTRAINT fk_grn_created_by FOREIGN KEY (created_by) REFERENCES users(id)
);

-- ============================================================================
-- LEGACY COMPATIBILITY TABLES
-- ============================================================================

-- Legacy Approval Tasks (includes document_number from migration 003)
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
    document_number VARCHAR(255), -- Added from migration 003
    approver_name VARCHAR(255),
    priority VARCHAR(50) DEFAULT 'medium',
    due_at TIMESTAMP,
    task_type VARCHAR(50) DEFAULT 'approval',
    title VARCHAR(255),
    workflow_id VARCHAR(255),
    workflow_name VARCHAR(255),
    stage_name VARCHAR(255),
    importance VARCHAR(50) DEFAULT 'medium',
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

-- Legacy Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id VARCHAR(255) NOT NULL,
    recipient_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    document_id VARCHAR(255),
    document_type VARCHAR(50),
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    sent BOOLEAN DEFAULT false,
    sent_at TIMESTAMP,
    entity_id VARCHAR(255),
    entity_type VARCHAR(50),
    entity_number VARCHAR(255),
    related_user_id VARCHAR(255),
    related_user_name VARCHAR(255),
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    action_taken BOOLEAN DEFAULT false,
    action_taken_at TIMESTAMP,
    importance VARCHAR(50),
    quick_action JSONB,
    reassignment_reason TEXT,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_notifications_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_notifications_recipient FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE
);
-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Core table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_organization ON users(current_organization_id);
CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);

-- Department indexes
CREATE INDEX IF NOT EXISTS idx_org_departments_organization ON organization_departments(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_departments_manager_name ON organization_departments(manager_name);

-- Workflow indexes
CREATE INDEX IF NOT EXISTS idx_workflows_organization ON workflows(organization_id);
CREATE INDEX IF NOT EXISTS idx_workflows_entity_type ON workflows(entity_type);
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_entity ON workflow_assignments(entity_id, entity_type);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_status ON workflow_tasks(status);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_assigned_role ON workflow_tasks(assigned_role);

-- Business document indexes
CREATE INDEX IF NOT EXISTS idx_requisitions_organization ON requisitions(organization_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_status ON requisitions(status);
CREATE INDEX IF NOT EXISTS idx_requisitions_document_number ON requisitions(document_number);
CREATE INDEX IF NOT EXISTS idx_budgets_organization ON budgets(organization_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_organization ON purchase_orders(organization_id);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_organization ON payment_vouchers(organization_id);
CREATE INDEX IF NOT EXISTS idx_grns_organization ON goods_received_notes(organization_id);

-- Approval task indexes
CREATE INDEX IF NOT EXISTS idx_approval_tasks_organization ON approval_tasks(organization_id);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_assigned_to ON approval_tasks(assigned_to);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_status ON approval_tasks(status);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_document_number ON approval_tasks(document_number);

-- Vendor and category indexes
CREATE INDEX IF NOT EXISTS idx_vendors_organization ON vendors(organization_id);
CREATE INDEX IF NOT EXISTS idx_categories_organization ON categories(organization_id);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_document ON audit_logs(document_id, document_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Notifications indexes
CREATE INDEX IF NOT EXISTS idx_notifications_organization ON notifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications(recipient_id);
CREATE INDEX IF NOT EXISTS idx_notifications_sent ON notifications(sent);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

-- ============================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- ============================================================================

-- Function to update updated_at timestamp
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

-- Documents table trigger
CREATE OR REPLACE FUNCTION update_documents_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_documents_updated_at();

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

-- Table comments
COMMENT ON TABLE workflows IS 'Enhanced workflow definitions with frontend compatibility';
COMMENT ON TABLE workflow_assignments IS 'Tracks workflow execution for specific entities';
COMMENT ON TABLE workflow_tasks IS 'Individual approval tasks within workflow assignments with concurrency control';
COMMENT ON TABLE stage_approval_records IS 'Tracks individual approvals per workflow stage for multiple approval support';
COMMENT ON TABLE task_assignment_history IS 'Tracks round-robin task assignment history for fair distribution';
COMMENT ON TABLE workflow_defaults IS 'Default workflow mappings for entity types per organization';
COMMENT ON TABLE documents IS 'Unified document table for all business document types';
COMMENT ON TABLE vendors IS 'Organization-scoped vendors for multi-tenant security';
COMMENT ON TABLE organization_departments IS 'Organization departments with manager name support';
COMMENT ON TABLE approval_tasks IS 'Legacy approval tasks with document number support';

-- ============================================================================
-- MIGRATION COMPLETION LOG
-- ============================================================================

-- Log successful completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 001_init_system completed successfully';
    RAISE NOTICE 'Created % tables with all business fields and enhancements', 
        (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public');
    RAISE NOTICE 'Created % indexes for performance optimization', 
        (SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public');
    RAISE NOTICE 'Database schema is ready for production use';
    RAISE NOTICE 'CONSOLIDATED FEATURES:';
    RAISE NOTICE '✅ Multi-tenant vendor isolation (organization_id)';
    RAISE NOTICE '✅ Unified documents table for search';
    RAISE NOTICE '✅ Nullable vendor_id in purchase_orders';
    RAISE NOTICE '✅ Manager name support in departments';
    RAISE NOTICE '✅ Document number support in approval tasks';
    RAISE NOTICE '✅ Enhanced workflow system with entity_type support';
    RAISE NOTICE '✅ Complete authentication and authorization system';
    RAISE NOTICE '✅ All automation fields and business enhancements';
END $$;