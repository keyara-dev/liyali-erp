-- NOTE: this will fail if multiple live (non-cancelled) PVs exist for the
-- same (linked_po, organization_id) — that is expected, since the whole
-- point of the up migration was to allow that. Resolve/cancel the extra PVs
-- before rolling back.
DROP INDEX IF EXISTS idx_pv_linked_po;

CREATE UNIQUE INDEX IF NOT EXISTS idx_pv_linked_po_unique
    ON payment_vouchers (linked_po, organization_id)
    WHERE UPPER(status) <> 'CANCELLED' AND linked_po <> '';

ALTER TABLE vendors DROP COLUMN IF EXISTS zra_tpin;
ALTER TABLE vendors DROP COLUMN IF EXISTS pacra_reg_number;
