-- Add missing automation fields to requisitions table
-- Migration: 003_add_automation_fields
-- Description: Adds automation_used and auto_created_po fields to requisitions table
-- Date: 2025-01-09

-- Add automation fields to requisitions table
ALTER TABLE requisitions 
ADD COLUMN IF NOT EXISTS automation_used BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS auto_created_po JSONB;

-- Add automation fields to purchase_orders table if not exists
ALTER TABLE purchase_orders 
ADD COLUMN IF NOT EXISTS automation_used BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS auto_created_grn JSONB;

-- Add automation fields to goods_received_notes table if not exists (should already exist)
ALTER TABLE goods_received_notes 
ADD COLUMN IF NOT EXISTS automation_used BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS auto_created_pv JSONB;

-- Add comments for documentation
COMMENT ON COLUMN requisitions.automation_used IS 'Whether automation was used in processing';
COMMENT ON COLUMN requisitions.auto_created_po IS 'Auto-created purchase order details';

COMMENT ON COLUMN purchase_orders.automation_used IS 'Whether automation was used in processing';
COMMENT ON COLUMN purchase_orders.auto_created_grn IS 'Auto-created GRN details';

SELECT 'Automation fields added successfully' as status;