-- Migration: Add fields for 100% frontend-backend alignment
-- This migration adds the remaining fields needed for perfect type alignment

-- Add Notes field to GRN items (stored in JSONB, no schema change needed for items array)
-- The Notes field will be part of the GRNItem JSON structure

-- Add extended fields to approval records (stored in JSONB, no schema change needed)
-- The extended fields will be part of the ApprovalRecord JSON structure

-- Update comments to reflect new status values and payment methods
COMMENT ON COLUMN goods_received_notes.status IS 'Status: draft, pending, approved, rejected, paid, completed, cancelled';
COMMENT ON COLUMN payment_vouchers.payment_method IS 'Payment method: bank_transfer, cash';

-- No actual schema changes needed as:
-- 1. GRNItem.Notes is stored in the items JSONB column
-- 2. ApprovalRecord extended fields are stored in the approval_history JSONB column
-- 3. Status and payment method enums are stored as strings with application-level validation