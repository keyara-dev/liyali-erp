-- Remove index
DROP INDEX IF EXISTS idx_requisitions_source_of_funds;

-- Remove column
ALTER TABLE requisitions DROP COLUMN IF EXISTS source_of_funds;
