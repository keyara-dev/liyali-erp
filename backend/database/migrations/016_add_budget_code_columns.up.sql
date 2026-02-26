-- Migration 016: Add budget_code to requisitions and budget fields to goods_received_notes
--
-- The requisitions table was created without a budget_code column, and the
-- goods_received_notes table lacks budget_code, cost_center, and project_code
-- columns entirely. This breaks the document chain where budget information
-- should flow: Requisition → PO → GRN → PV.

-- Add budget_code column to requisitions (cost_center and project_code already exist)
ALTER TABLE requisitions ADD COLUMN IF NOT EXISTS budget_code VARCHAR(255);

-- Add budget tracking columns to goods_received_notes
ALTER TABLE goods_received_notes ADD COLUMN IF NOT EXISTS budget_code VARCHAR(255);
ALTER TABLE goods_received_notes ADD COLUMN IF NOT EXISTS cost_center VARCHAR(255);
ALTER TABLE goods_received_notes ADD COLUMN IF NOT EXISTS project_code VARCHAR(255);
