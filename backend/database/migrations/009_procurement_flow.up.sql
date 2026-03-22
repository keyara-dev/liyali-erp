-- Add procurement flow support
-- Goods-First flow: PV references the approved GRN that preceded it
-- Payment-First flow: GRN references the approved PV that preceded it

-- Cross-link columns
ALTER TABLE payment_vouchers ADD COLUMN IF NOT EXISTS linked_grn VARCHAR(100) DEFAULT '';
ALTER TABLE goods_received_notes ADD COLUMN IF NOT EXISTS linked_pv VARCHAR(100) DEFAULT '';

-- Per-PO override (empty = inherit from org setting)
ALTER TABLE purchase_orders ADD COLUMN IF NOT EXISTS procurement_flow VARCHAR(20) DEFAULT '';

-- Org-level default
ALTER TABLE organization_settings ADD COLUMN IF NOT EXISTS procurement_flow VARCHAR(20) DEFAULT 'goods_first';
