-- Group 1: Supporting Documents on PO
ALTER TABLE purchase_orders ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Group 2: Cost tracking
ALTER TABLE purchase_orders ADD COLUMN IF NOT EXISTS estimated_cost NUMERIC(15,2) NOT NULL DEFAULT 0;

-- Group 4: Quotation bypass
ALTER TABLE purchase_orders ADD COLUMN IF NOT EXISTS quotation_gate_overridden BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE purchase_orders ADD COLUMN IF NOT EXISTS bypass_justification TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_purchase_orders_estimated_cost ON purchase_orders (estimated_cost) WHERE estimated_cost > 0;

-- Group 5: Vendor model expansion
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS bank_name       VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS account_name    VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS account_number  VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS branch_code     VARCHAR(50)  NOT NULL DEFAULT '';
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS swift_code      VARCHAR(20)  NOT NULL DEFAULT '';
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS contact_person  VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS physical_address TEXT         NOT NULL DEFAULT '';
