-- ============================================================================
-- PO DELIVERY TRACKING
-- Migration: 019_po_delivery_tracking
-- Adds delivery_status to purchase_orders. Independent of workflow Status —
-- tracks PHYSICAL receipt as GRNs are approved against the PO.
-- Per-item received quantity lives in the items JSONB blob (no schema change).
-- ============================================================================

ALTER TABLE purchase_orders
    ADD COLUMN IF NOT EXISTS delivery_status TEXT NOT NULL DEFAULT 'NOT_DELIVERED';

CREATE INDEX IF NOT EXISTS idx_po_delivery_status
    ON purchase_orders (organization_id, delivery_status);
