-- Rollback migration: Remove alignment fields
-- This rollback removes the comments added for alignment

-- Remove updated comments
COMMENT ON COLUMN goods_received_notes.status IS 'Status: draft, pending, approved, rejected, completed, cancelled';
COMMENT ON COLUMN payment_vouchers.payment_method IS 'Payment method: bank_transfer, check, cash';

-- Note: No actual schema rollback needed as the changes were:
-- 1. Adding optional fields to JSONB structures (backward compatible)
-- 2. Updating enum value comments (no breaking changes)