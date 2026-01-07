-- ============================================================================
-- LIYALI GATEWAY - ROLLBACK MISSING FIELDS MIGRATION
-- Migration: 002_add_missing_fields (DOWN)
-- Description: Removes all fields added in the up migration
-- ============================================================================

-- ============================================================================
-- DROP FOREIGN KEY CONSTRAINTS
-- ============================================================================

ALTER TABLE requisitions DROP CONSTRAINT IF EXISTS fk_requisitions_created_by;
ALTER TABLE budgets DROP CONSTRAINT IF EXISTS fk_budgets_created_by;
ALTER TABLE purchase_orders DROP CONSTRAINT IF EXISTS fk_purchase_orders_created_by;
ALTER TABLE purchase_orders DROP CONSTRAINT IF EXISTS fk_purchase_orders_source_requisition_id;
ALTER TABLE payment_vouchers DROP CONSTRAINT IF EXISTS fk_payment_vouchers_created_by;
ALTER TABLE goods_received_notes DROP CONSTRAINT IF EXISTS fk_grn_created_by;

-- ============================================================================
-- DROP INDEXES
-- ============================================================================

DROP INDEX IF EXISTS idx_requisitions_department_id;
DROP INDEX IF EXISTS idx_requisitions_created_by;
DROP INDEX IF EXISTS idx_requisitions_budget_code;
DROP INDEX IF EXISTS idx_requisitions_cost_center;

DROP INDEX IF EXISTS idx_budgets_department_id;
DROP INDEX IF EXISTS idx_budgets_created_by;

DROP INDEX IF EXISTS idx_purchase_orders_department_id;
DROP INDEX IF EXISTS idx_purchase_orders_created_by;
DROP INDEX IF EXISTS idx_purchase_orders_budget_code;
DROP INDEX IF EXISTS idx_purchase_orders_cost_center;
DROP INDEX IF EXISTS idx_purchase_orders_source_requisition_id;

DROP INDEX IF EXISTS idx_payment_vouchers_department_id;
DROP INDEX IF EXISTS idx_payment_vouchers_created_by;
DROP INDEX IF EXISTS idx_payment_vouchers_budget_code;
DROP INDEX IF EXISTS idx_payment_vouchers_cost_center;
DROP INDEX IF EXISTS idx_payment_vouchers_payment_due_date;

DROP INDEX IF EXISTS idx_grn_created_by;
DROP INDEX IF EXISTS idx_grn_warehouse_location;

DROP INDEX IF EXISTS idx_approval_tasks_priority;
DROP INDEX IF EXISTS idx_approval_tasks_due_at;
DROP INDEX IF EXISTS idx_approval_tasks_workflow_id;

-- ============================================================================
-- REQUISITIONS TABLE - Remove added fields
-- ============================================================================

ALTER TABLE requisitions 
DROP COLUMN IF EXISTS department_id,
DROP COLUMN IF EXISTS required_by_date,
DROP COLUMN IF EXISTS cost_center,
DROP COLUMN IF EXISTS project_code,
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS created_by_role,
DROP COLUMN IF EXISTS metadata;

-- ============================================================================
-- BUDGETS TABLE - Remove added fields
-- ============================================================================

ALTER TABLE budgets 
DROP COLUMN IF EXISTS name,
DROP COLUMN IF EXISTS description,
DROP COLUMN IF EXISTS department_id,
DROP COLUMN IF EXISTS currency,
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS items,
DROP COLUMN IF EXISTS action_history,
DROP COLUMN IF EXISTS metadata;

-- ============================================================================
-- PURCHASE_ORDERS TABLE - Remove added fields
-- ============================================================================

ALTER TABLE purchase_orders 
DROP COLUMN IF EXISTS description,
DROP COLUMN IF EXISTS department,
DROP COLUMN IF EXISTS department_id,
DROP COLUMN IF EXISTS gl_code,
DROP COLUMN IF EXISTS title,
DROP COLUMN IF EXISTS priority,
DROP COLUMN IF EXISTS subtotal,
DROP COLUMN IF EXISTS tax,
DROP COLUMN IF EXISTS total,
DROP COLUMN IF EXISTS budget_code,
DROP COLUMN IF EXISTS cost_center,
DROP COLUMN IF EXISTS project_code,
DROP COLUMN IF EXISTS required_by_date,
DROP COLUMN IF EXISTS source_requisition_number,
DROP COLUMN IF EXISTS source_requisition_id,
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS owner_id,
DROP COLUMN IF EXISTS action_history,
DROP COLUMN IF EXISTS metadata;

-- ============================================================================
-- PAYMENT_VOUCHERS TABLE - Remove added fields
-- ============================================================================

ALTER TABLE payment_vouchers 
DROP COLUMN IF EXISTS title,
DROP COLUMN IF EXISTS department,
DROP COLUMN IF EXISTS department_id,
DROP COLUMN IF EXISTS priority,
DROP COLUMN IF EXISTS requested_by_name,
DROP COLUMN IF EXISTS requested_date,
DROP COLUMN IF EXISTS submitted_at,
DROP COLUMN IF EXISTS approved_at,
DROP COLUMN IF EXISTS paid_date,
DROP COLUMN IF EXISTS payment_due_date,
DROP COLUMN IF EXISTS budget_code,
DROP COLUMN IF EXISTS cost_center,
DROP COLUMN IF EXISTS project_code,
DROP COLUMN IF EXISTS tax_amount,
DROP COLUMN IF EXISTS withholding_tax_amount,
DROP COLUMN IF EXISTS paid_amount,
DROP COLUMN IF EXISTS source_purchase_order_number,
DROP COLUMN IF EXISTS source_requisition_number,
DROP COLUMN IF EXISTS bank_details,
DROP COLUMN IF EXISTS items,
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS owner_id,
DROP COLUMN IF EXISTS action_history,
DROP COLUMN IF EXISTS metadata;

-- ============================================================================
-- GOODS_RECEIVED_NOTES TABLE - Remove added fields
-- ============================================================================

ALTER TABLE goods_received_notes 
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS owner_id,
DROP COLUMN IF EXISTS warehouse_location,
DROP COLUMN IF EXISTS notes,
DROP COLUMN IF EXISTS stage_name,
DROP COLUMN IF EXISTS approved_by,
DROP COLUMN IF EXISTS automation_used,
DROP COLUMN IF EXISTS auto_created_pv,
DROP COLUMN IF EXISTS action_history,
DROP COLUMN IF EXISTS metadata;

-- ============================================================================
-- APPROVAL_TASKS TABLE - Remove added fields
-- ============================================================================

ALTER TABLE approval_tasks 
DROP COLUMN IF EXISTS priority,
DROP COLUMN IF EXISTS due_at,
DROP COLUMN IF EXISTS task_type,
DROP COLUMN IF EXISTS title,
DROP COLUMN IF EXISTS workflow_id,
DROP COLUMN IF EXISTS workflow_name,
DROP COLUMN IF EXISTS stage_name,
DROP COLUMN IF EXISTS importance;