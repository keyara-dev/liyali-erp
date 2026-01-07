-- ============================================================================
-- LIYALI GATEWAY - ADD MISSING FIELDS FOR TYPE ALIGNMENT
-- Migration: 002_add_missing_fields
-- Description: Adds all missing fields to align backend models with frontend types
-- ============================================================================

-- ============================================================================
-- REQUISITIONS TABLE - Add missing fields
-- ============================================================================

ALTER TABLE requisitions 
ADD COLUMN IF NOT EXISTS department_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS required_by_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS cost_center VARCHAR(255),
ADD COLUMN IF NOT EXISTS project_code VARCHAR(255),
ADD COLUMN IF NOT EXISTS created_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS created_by_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS created_by_role VARCHAR(255),
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- ============================================================================
-- BUDGETS TABLE - Add missing fields
-- ============================================================================

ALTER TABLE budgets 
ADD COLUMN IF NOT EXISTS name VARCHAR(255),
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS department_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'USD',
ADD COLUMN IF NOT EXISTS created_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS items JSONB,
ADD COLUMN IF NOT EXISTS action_history JSONB,
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- ============================================================================
-- PURCHASE_ORDERS TABLE - Add missing fields
-- ============================================================================

ALTER TABLE purchase_orders 
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS department VARCHAR(255),
ADD COLUMN IF NOT EXISTS department_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS gl_code VARCHAR(255),
ADD COLUMN IF NOT EXISTS title VARCHAR(255),
ADD COLUMN IF NOT EXISTS priority VARCHAR(50) DEFAULT 'medium',
ADD COLUMN IF NOT EXISTS subtotal DECIMAL(15,2),
ADD COLUMN IF NOT EXISTS tax DECIMAL(15,2),
ADD COLUMN IF NOT EXISTS total DECIMAL(15,2),
ADD COLUMN IF NOT EXISTS budget_code VARCHAR(255),
ADD COLUMN IF NOT EXISTS cost_center VARCHAR(255),
ADD COLUMN IF NOT EXISTS project_code VARCHAR(255),
ADD COLUMN IF NOT EXISTS required_by_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS source_requisition_number VARCHAR(255),
ADD COLUMN IF NOT EXISTS source_requisition_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS created_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS owner_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS action_history JSONB,
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- ============================================================================
-- PAYMENT_VOUCHERS TABLE - Add missing fields
-- ============================================================================

ALTER TABLE payment_vouchers 
ADD COLUMN IF NOT EXISTS title VARCHAR(255),
ADD COLUMN IF NOT EXISTS department VARCHAR(255),
ADD COLUMN IF NOT EXISTS department_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS priority VARCHAR(50) DEFAULT 'medium',
ADD COLUMN IF NOT EXISTS requested_by_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS requested_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS paid_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS payment_due_date TIMESTAMP,
ADD COLUMN IF NOT EXISTS budget_code VARCHAR(255),
ADD COLUMN IF NOT EXISTS cost_center VARCHAR(255),
ADD COLUMN IF NOT EXISTS project_code VARCHAR(255),
ADD COLUMN IF NOT EXISTS tax_amount DECIMAL(15,2),
ADD COLUMN IF NOT EXISTS withholding_tax_amount DECIMAL(15,2),
ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(15,2),
ADD COLUMN IF NOT EXISTS source_purchase_order_number VARCHAR(255),
ADD COLUMN IF NOT EXISTS source_requisition_number VARCHAR(255),
ADD COLUMN IF NOT EXISTS bank_details JSONB,
ADD COLUMN IF NOT EXISTS items JSONB,
ADD COLUMN IF NOT EXISTS created_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS owner_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS action_history JSONB,
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- ============================================================================
-- GOODS_RECEIVED_NOTES TABLE - Add missing fields
-- ============================================================================

ALTER TABLE goods_received_notes 
ADD COLUMN IF NOT EXISTS created_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS owner_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS warehouse_location VARCHAR(255),
ADD COLUMN IF NOT EXISTS notes TEXT,
ADD COLUMN IF NOT EXISTS stage_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS approved_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS automation_used BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS auto_created_pv JSONB,
ADD COLUMN IF NOT EXISTS action_history JSONB,
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- ============================================================================
-- APPROVAL_TASKS TABLE - Add missing fields
-- ============================================================================

ALTER TABLE approval_tasks 
ADD COLUMN IF NOT EXISTS priority VARCHAR(50) DEFAULT 'medium',
ADD COLUMN IF NOT EXISTS due_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS task_type VARCHAR(100),
ADD COLUMN IF NOT EXISTS title VARCHAR(255),
ADD COLUMN IF NOT EXISTS workflow_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS workflow_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS stage_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS importance VARCHAR(50) DEFAULT 'medium';

-- ============================================================================
-- ADD INDEXES FOR PERFORMANCE
-- ============================================================================

-- Requisitions indexes
CREATE INDEX IF NOT EXISTS idx_requisitions_department_id ON requisitions(department_id);
CREATE INDEX IF NOT EXISTS idx_requisitions_created_by ON requisitions(created_by);
CREATE INDEX IF NOT EXISTS idx_requisitions_budget_code ON requisitions(budget_code);
CREATE INDEX IF NOT EXISTS idx_requisitions_cost_center ON requisitions(cost_center);

-- Budgets indexes
CREATE INDEX IF NOT EXISTS idx_budgets_department_id ON budgets(department_id);
CREATE INDEX IF NOT EXISTS idx_budgets_created_by ON budgets(created_by);

-- Purchase Orders indexes
CREATE INDEX IF NOT EXISTS idx_purchase_orders_department_id ON purchase_orders(department_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_created_by ON purchase_orders(created_by);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_budget_code ON purchase_orders(budget_code);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_cost_center ON purchase_orders(cost_center);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_source_requisition_id ON purchase_orders(source_requisition_id);

-- Payment Vouchers indexes
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_department_id ON payment_vouchers(department_id);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_created_by ON payment_vouchers(created_by);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_budget_code ON payment_vouchers(budget_code);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_cost_center ON payment_vouchers(cost_center);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_payment_due_date ON payment_vouchers(payment_due_date);

-- GRN indexes
CREATE INDEX IF NOT EXISTS idx_grn_created_by ON goods_received_notes(created_by);
CREATE INDEX IF NOT EXISTS idx_grn_warehouse_location ON goods_received_notes(warehouse_location);

-- Approval Tasks indexes
CREATE INDEX IF NOT EXISTS idx_approval_tasks_priority ON approval_tasks(priority);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_due_at ON approval_tasks(due_at);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_workflow_id ON approval_tasks(workflow_id);

-- ============================================================================
-- ADD FOREIGN KEY CONSTRAINTS
-- ============================================================================

-- Requisitions foreign keys
ALTER TABLE requisitions 
ADD CONSTRAINT IF NOT EXISTS fk_requisitions_created_by 
FOREIGN KEY (created_by) REFERENCES users(id);

-- Budgets foreign keys
ALTER TABLE budgets 
ADD CONSTRAINT IF NOT EXISTS fk_budgets_created_by 
FOREIGN KEY (created_by) REFERENCES users(id);

-- Purchase Orders foreign keys
ALTER TABLE purchase_orders 
ADD CONSTRAINT IF NOT EXISTS fk_purchase_orders_created_by 
FOREIGN KEY (created_by) REFERENCES users(id),
ADD CONSTRAINT IF NOT EXISTS fk_purchase_orders_source_requisition_id 
FOREIGN KEY (source_requisition_id) REFERENCES requisitions(id);

-- Payment Vouchers foreign keys
ALTER TABLE payment_vouchers 
ADD CONSTRAINT IF NOT EXISTS fk_payment_vouchers_created_by 
FOREIGN KEY (created_by) REFERENCES users(id);

-- GRN foreign keys
ALTER TABLE goods_received_notes 
ADD CONSTRAINT IF NOT EXISTS fk_grn_created_by 
FOREIGN KEY (created_by) REFERENCES users(id);

-- ============================================================================
-- UPDATE EXISTING DATA (Set defaults for new fields)
-- ============================================================================

-- Set default currency for existing budgets
UPDATE budgets SET currency = 'USD' WHERE currency IS NULL;

-- Set default priority for existing records
UPDATE requisitions SET priority = 'medium' WHERE priority IS NULL;
UPDATE purchase_orders SET priority = 'medium' WHERE priority IS NULL;
UPDATE payment_vouchers SET priority = 'medium' WHERE priority IS NULL;

-- Set automation_used to false for existing GRNs
UPDATE goods_received_notes SET automation_used = FALSE WHERE automation_used IS NULL;

-- Set default importance for existing approval tasks
UPDATE approval_tasks SET importance = 'medium' WHERE importance IS NULL;

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON COLUMN requisitions.department_id IS 'Department ID reference';
COMMENT ON COLUMN requisitions.budget_code IS 'Budget code for tracking';
COMMENT ON COLUMN requisitions.cost_center IS 'Cost center for accounting';
COMMENT ON COLUMN requisitions.project_code IS 'Project code for tracking';
COMMENT ON COLUMN requisitions.metadata IS 'Generic metadata for extensibility';

COMMENT ON COLUMN budgets.name IS 'Budget display name';
COMMENT ON COLUMN budgets.items IS 'Budget line items breakdown';
COMMENT ON COLUMN budgets.metadata IS 'Generic metadata for extensibility';

COMMENT ON COLUMN purchase_orders.gl_code IS 'General Ledger code';
COMMENT ON COLUMN purchase_orders.subtotal IS 'Subtotal before tax';
COMMENT ON COLUMN purchase_orders.tax IS 'Tax amount';
COMMENT ON COLUMN purchase_orders.total IS 'Total amount including tax';
COMMENT ON COLUMN purchase_orders.metadata IS 'Generic metadata for extensibility';

COMMENT ON COLUMN payment_vouchers.tax_amount IS 'Tax amount for payment';
COMMENT ON COLUMN payment_vouchers.withholding_tax_amount IS 'Withholding tax amount';
COMMENT ON COLUMN payment_vouchers.paid_amount IS 'Actual amount paid';
COMMENT ON COLUMN payment_vouchers.bank_details IS 'Bank details for payment';
COMMENT ON COLUMN payment_vouchers.items IS 'Payment line items breakdown';
COMMENT ON COLUMN payment_vouchers.metadata IS 'Generic metadata for extensibility';

COMMENT ON COLUMN goods_received_notes.automation_used IS 'Whether automation was used in processing';
COMMENT ON COLUMN goods_received_notes.auto_created_pv IS 'Auto-created payment voucher details';
COMMENT ON COLUMN goods_received_notes.metadata IS 'Generic metadata for extensibility';