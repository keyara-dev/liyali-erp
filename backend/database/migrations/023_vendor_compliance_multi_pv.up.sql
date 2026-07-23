-- Item A: supplier compliance fields (pattern: 011_po_metadata_estimated_cost.up.sql:14-19)
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS zra_tpin         VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS pacra_reg_number VARCHAR(100) NOT NULL DEFAULT '';
-- Legacy tax_id was captured as "Tax ID / TPIN" in the UI — seed the dedicated column
UPDATE vendors SET zra_tpin = tax_id WHERE zra_tpin = '' AND tax_id <> '';

-- Item B: multiple PVs per PO — drop one-live-PV unique index (013:10-12), replace with lookup index
DROP INDEX IF EXISTS idx_pv_linked_po_unique;
CREATE INDEX IF NOT EXISTS idx_pv_linked_po ON payment_vouchers (linked_po, organization_id) WHERE linked_po <> '';
