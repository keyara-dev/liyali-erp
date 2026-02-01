-- Remove action_history column from requisitions table

ALTER TABLE requisitions DROP COLUMN IF EXISTS action_history;
