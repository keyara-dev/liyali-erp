-- Rollback migration 014_add_reports_indexes

DROP INDEX IF EXISTS idx_requisitions_org_status_created;
DROP INDEX IF EXISTS idx_purchase_orders_org_status_created;
DROP INDEX IF EXISTS idx_payment_vouchers_org_status_created;
DROP INDEX IF EXISTS idx_grn_org_status_created;
DROP INDEX IF EXISTS idx_budgets_org_status_created;
DROP INDEX IF EXISTS idx_stage_approval_org_created;
DROP INDEX IF EXISTS idx_stage_approval_org_action;
DROP INDEX IF EXISTS idx_workflow_tasks_org_completed;
DROP INDEX IF EXISTS idx_workflow_tasks_stage_times;
