-- Rollback Migration 016: Remove budget_code columns

ALTER TABLE requisitions DROP COLUMN IF EXISTS budget_code;

ALTER TABLE goods_received_notes DROP COLUMN IF EXISTS budget_code;
ALTER TABLE goods_received_notes DROP COLUMN IF EXISTS cost_center;
ALTER TABLE goods_received_notes DROP COLUMN IF EXISTS project_code;
