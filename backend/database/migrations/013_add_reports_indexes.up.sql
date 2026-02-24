-- ============================================================================
-- ADMIN REPORTS PERFORMANCE INDEXES
-- Migration: 014_add_reports_indexes
-- Description: Add indexes to optimize admin reports and analytics queries
-- Date: 2026-02-22
-- ============================================================================

-- For date range filtering and sorting on documents
CREATE INDEX IF NOT EXISTS idx_requisitions_org_status_created 
ON requisitions(organization_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_purchase_orders_org_status_created 
ON purchase_orders(organization_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_vouchers_org_status_created 
ON payment_vouchers(organization_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_grn_org_status_created 
ON goods_received_notes(organization_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_budgets_org_status_created 
ON budgets(organization_id, status, created_at DESC);

-- For approval activity queries
CREATE INDEX IF NOT EXISTS idx_stage_approval_org_created 
ON stage_approval_records(organization_id, approved_at DESC);

CREATE INDEX IF NOT EXISTS idx_stage_approval_org_action 
ON stage_approval_records(organization_id, action, approved_at DESC);

-- For workflow task performance queries
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_org_completed 
ON workflow_tasks(organization_id, completed_at DESC) 
WHERE completed_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_workflow_tasks_stage_times 
ON workflow_tasks(organization_id, stage_number, created_at, completed_at) 
WHERE completed_at IS NOT NULL;

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 014_add_reports_indexes completed successfully';
    RAISE NOTICE 'Added 9 performance indexes for admin reports';
END $$;
