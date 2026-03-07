-- Add source_of_funds field to requisitions table
ALTER TABLE requisitions ADD COLUMN IF NOT EXISTS source_of_funds VARCHAR(255);

-- Add index for filtering/reporting
CREATE INDEX IF NOT EXISTS idx_requisitions_source_of_funds ON requisitions(source_of_funds);
