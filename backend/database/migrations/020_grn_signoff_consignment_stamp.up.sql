-- ============================================================================
-- GRN sign-off (Receiver + Certifier), consignment note, vendor snapshot,
-- and organization stamp image.
-- Migration: 020_grn_signoff_consignment_stamp
--
-- Matches the printed GRN form used by Zambian councils: Supplier Name +
-- Address block, Delivery Consignment Note, per-line Item Code + Remarks,
-- separate Received By / Certified By signature blocks and Stamp of
-- Issuing Officer.
-- ============================================================================

ALTER TABLE goods_received_notes
    -- Delivery / consignment note number printed in the PDF header
    ADD COLUMN IF NOT EXISTS consignment_note TEXT NOT NULL DEFAULT '',

    -- Vendor snapshot captured at GRN creation time (audit-safe; later vendor
    -- renames / deletions do not mutate historical PDFs).
    ADD COLUMN IF NOT EXISTS vendor_name    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS vendor_address TEXT NOT NULL DEFAULT '',

    -- Receiver sign-off (the person who physically received the goods)
    ADD COLUMN IF NOT EXISTS received_by_name      TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS received_by_signature TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS received_at           TIMESTAMPTZ,

    -- Certifier sign-off (issuing officer who certifies the receipt)
    ADD COLUMN IF NOT EXISTS certified_by_id        TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS certified_by_name      TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS certified_by_signature TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS certified_at           TIMESTAMPTZ,

    -- Sign-off state machine, independent of workflow Status:
    --   PENDING_RECEIVER  -> PENDING_CERTIFIER -> READY -> COMPLETED
    -- Workflow submission is allowed only when signoff_status = 'READY'.
    ADD COLUMN IF NOT EXISTS signoff_status TEXT NOT NULL DEFAULT 'PENDING_RECEIVER',

    -- Per-GRN stamp image, captured at certification time. Falls back to
    -- organization_settings.stamp_image_url on the PDF when empty.
    ADD COLUMN IF NOT EXISTS stamp_image_url TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_grn_signoff_status
    ON goods_received_notes (organization_id, signoff_status);

-- Organization-level stamp image (single "Stamp of Issuing Officer" graphic
-- printed on every GRN PDF). Stored as URL — actual upload handled by the
-- existing file-upload pipeline.
ALTER TABLE organization_settings
    ADD COLUMN IF NOT EXISTS stamp_image_url TEXT NOT NULL DEFAULT '';
