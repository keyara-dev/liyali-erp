-- Migration 019: Add deleted_at to document tables
--
-- The getCurrentUsage limit-check middleware queries all document tables with
-- "AND deleted_at IS NULL", but several tables were created without this column.
-- Adding it (defaulting to NULL) makes all existing rows visible to the query,
-- and keeps the door open for soft-delete if needed in future.

ALTER TABLE budgets             ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE requisitions        ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE purchase_orders     ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE payment_vouchers    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE goods_received_notes ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE vendors             ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Partial indexes for efficient soft-delete filtering
CREATE INDEX IF NOT EXISTS idx_budgets_deleted_at              ON budgets(deleted_at)              WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_requisitions_deleted_at         ON requisitions(deleted_at)         WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_purchase_orders_deleted_at      ON purchase_orders(deleted_at)      WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_deleted_at     ON payment_vouchers(deleted_at)     WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_grn_deleted_at                  ON goods_received_notes(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vendors_deleted_at              ON vendors(deleted_at)              WHERE deleted_at IS NULL;
