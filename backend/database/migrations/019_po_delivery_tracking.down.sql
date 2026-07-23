DROP INDEX IF EXISTS idx_po_delivery_status;
ALTER TABLE purchase_orders DROP COLUMN IF EXISTS delivery_status;
