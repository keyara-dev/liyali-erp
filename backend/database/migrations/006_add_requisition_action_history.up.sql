-- Add action_history column to requisitions table
-- This column stores the timeline of all actions taken on a requisition

ALTER TABLE requisitions ADD COLUMN IF NOT EXISTS action_history JSONB;

-- Add a comment for documentation
COMMENT ON COLUMN requisitions.action_history IS 'JSON array of action history entries for timeline display';
