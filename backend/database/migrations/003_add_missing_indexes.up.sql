-- ============================================================================
-- LIYALI GATEWAY - ADD MISSING INDEXES
-- Migration: 003_add_missing_indexes
-- Description: Add indexes for commonly queried foreign keys and columns
-- Date: January 21, 2026
-- ============================================================================

-- Organization members indexes (frequently queried by org_id + user_id)
CREATE INDEX IF NOT EXISTS idx_organization_members_org_id ON organization_members(organization_id);
CREATE INDEX IF NOT EXISTS idx_organization_members_user_id ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_organization_members_org_user ON organization_members(organization_id, user_id);

-- Sessions indexes (queried by user_id for session management)
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Workflow tasks assignment index (frequently queried)
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_assignment_id ON workflow_tasks(workflow_assignment_id);

-- Login attempts indexes (queried for rate limiting and security)
CREATE INDEX IF NOT EXISTS idx_login_attempts_email ON login_attempts(email);
CREATE INDEX IF NOT EXISTS idx_login_attempts_attempted_at ON login_attempts(attempted_at);
CREATE INDEX IF NOT EXISTS idx_login_attempts_email_time ON login_attempts(email, attempted_at);

-- Account lockouts indexes
CREATE INDEX IF NOT EXISTS idx_account_lockouts_user_id ON account_lockouts(user_id);
CREATE INDEX IF NOT EXISTS idx_account_lockouts_active ON account_lockouts(user_id, active);

-- Password resets indexes
CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);

-- User organization roles indexes
CREATE INDEX IF NOT EXISTS idx_user_org_roles_user_id ON user_organization_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_org_roles_org_id ON user_organization_roles(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_org_roles_role_id ON user_organization_roles(role_id);

-- Stage approval records indexes
CREATE INDEX IF NOT EXISTS idx_stage_approval_task_id ON stage_approval_records(workflow_task_id);
CREATE INDEX IF NOT EXISTS idx_stage_approval_approver_id ON stage_approval_records(approver_id);

-- Documents indexes for search
CREATE INDEX IF NOT EXISTS idx_documents_organization ON documents(organization_id);
CREATE INDEX IF NOT EXISTS idx_documents_type ON documents(document_type);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_created_by ON documents(created_by);
CREATE INDEX IF NOT EXISTS idx_documents_number ON documents(document_number);

-- Workflow assignments composite index
CREATE INDEX IF NOT EXISTS idx_workflow_assignments_org_entity ON workflow_assignments(organization_id, entity_type);

-- Requisitions additional indexes
CREATE INDEX IF NOT EXISTS idx_requisitions_requester ON requisitions(requester_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_category ON requisitions(category_id);

-- Purchase orders additional indexes
CREATE INDEX IF NOT EXISTS idx_purchase_orders_vendor ON purchase_orders(vendor_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_status ON purchase_orders(status);

-- Payment vouchers additional indexes
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_vendor ON payment_vouchers(vendor_id);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_status ON payment_vouchers(status);

-- GRN additional indexes
CREATE INDEX IF NOT EXISTS idx_grns_status ON goods_received_notes(status);
CREATE INDEX IF NOT EXISTS idx_grns_received_by ON goods_received_notes(received_by);

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 003_add_missing_indexes completed successfully';
    RAISE NOTICE 'Added indexes for frequently queried foreign keys and columns';
END $$;
