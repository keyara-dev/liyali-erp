-- Migration 018: Add tagline column to organizations
--
-- Adds a tagline field to the organizations table so each workspace can
-- configure a sub-heading that appears on all generated PDF documents
-- (Requisition, Purchase Order, Payment Voucher, Goods Received Note).

ALTER TABLE organizations ADD COLUMN IF NOT EXISTS tagline VARCHAR(500);
