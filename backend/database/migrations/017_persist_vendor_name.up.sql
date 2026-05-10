-- ============================================================================
-- PERSIST MANUAL VENDOR NAME
-- Migration: 017_persist_vendor_name
-- Adds a nullable vendor_name / preferred_vendor_name column to documents
-- so quotations with manually-typed (non-system) vendors can be selected
-- as the supplier without losing the vendor name on save.
-- ============================================================================

ALTER TABLE purchase_orders   ADD COLUMN IF NOT EXISTS vendor_name           VARCHAR(255);
ALTER TABLE payment_vouchers  ADD COLUMN IF NOT EXISTS vendor_name           VARCHAR(255);
ALTER TABLE requisitions      ADD COLUMN IF NOT EXISTS preferred_vendor_name VARCHAR(255);
