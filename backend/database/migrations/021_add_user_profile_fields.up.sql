-- Add position, man_number, nrc_number, and contact fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS position VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS man_number VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS nrc_number VARCHAR(100);
ALTER TABLE users ADD COLUMN IF NOT EXISTS contact VARCHAR(50);

-- Add indexes for searchability
CREATE INDEX IF NOT EXISTS idx_users_man_number ON users(man_number);
CREATE INDEX IF NOT EXISTS idx_users_nrc_number ON users(nrc_number);
