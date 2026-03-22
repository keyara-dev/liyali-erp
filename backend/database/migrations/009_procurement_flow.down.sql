ALTER TABLE payment_vouchers DROP COLUMN IF EXISTS linked_grn;
ALTER TABLE goods_received_notes DROP COLUMN IF EXISTS linked_pv;
ALTER TABLE purchase_orders DROP COLUMN IF EXISTS procurement_flow;
ALTER TABLE organization_settings DROP COLUMN IF EXISTS procurement_flow;
