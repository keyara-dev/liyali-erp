-- Rollback Migration 019
DROP INDEX IF EXISTS idx_budgets_deleted_at;
DROP INDEX IF EXISTS idx_requisitions_deleted_at;
DROP INDEX IF EXISTS idx_purchase_orders_deleted_at;
DROP INDEX IF EXISTS idx_payment_vouchers_deleted_at;
DROP INDEX IF EXISTS idx_grn_deleted_at;
DROP INDEX IF EXISTS idx_vendors_deleted_at;

ALTER TABLE budgets             DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE requisitions        DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE purchase_orders     DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE payment_vouchers    DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE goods_received_notes DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE vendors             DROP COLUMN IF EXISTS deleted_at;
