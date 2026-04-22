-- One-to-one link enforcement between PO, GRN, and PV.
--
-- Application-level "check-then-insert" guards are racy: two concurrent creates
-- for the same PO can both pass the existence check and both insert. Partial
-- unique indexes close the race at the database layer while still allowing
-- cancelled rows to exist alongside a live one (so a mistaken doc can be
-- cancelled and re-created).

-- PV: exactly one non-cancelled PV per (linked_po, organization_id).
CREATE UNIQUE INDEX IF NOT EXISTS idx_pv_linked_po_unique
    ON payment_vouchers (linked_po, organization_id)
    WHERE UPPER(status) <> 'CANCELLED' AND linked_po <> '';

-- GRN goods-first: exactly one non-cancelled GRN per (po_document_number, organization_id)
-- — but only when linked_pv is empty (goods-first path). Payment-first GRNs
-- are uniqued by linked_pv in the index below.
CREATE UNIQUE INDEX IF NOT EXISTS idx_grn_po_unique_goods_first
    ON goods_received_notes (po_document_number, organization_id)
    WHERE UPPER(status) <> 'CANCELLED' AND linked_pv = '';

-- GRN payment-first: exactly one non-cancelled GRN per (linked_pv, organization_id).
CREATE UNIQUE INDEX IF NOT EXISTS idx_grn_pv_unique_payment_first
    ON goods_received_notes (linked_pv, organization_id)
    WHERE UPPER(status) <> 'CANCELLED' AND linked_pv <> '';
