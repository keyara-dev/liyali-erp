-- Reverse of 020_grn_signoff_consignment_stamp

DROP INDEX IF EXISTS idx_grn_signoff_status;

ALTER TABLE goods_received_notes
    DROP COLUMN IF EXISTS consignment_note,
    DROP COLUMN IF EXISTS vendor_name,
    DROP COLUMN IF EXISTS vendor_address,
    DROP COLUMN IF EXISTS received_by_name,
    DROP COLUMN IF EXISTS received_by_signature,
    DROP COLUMN IF EXISTS received_at,
    DROP COLUMN IF EXISTS certified_by_id,
    DROP COLUMN IF EXISTS certified_by_name,
    DROP COLUMN IF EXISTS certified_by_signature,
    DROP COLUMN IF EXISTS certified_at,
    DROP COLUMN IF EXISTS signoff_status,
    DROP COLUMN IF EXISTS stamp_image_url;

ALTER TABLE organization_settings
    DROP COLUMN IF EXISTS stamp_image_url;
