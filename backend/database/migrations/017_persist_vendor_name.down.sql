ALTER TABLE purchase_orders   DROP COLUMN IF EXISTS vendor_name;
ALTER TABLE payment_vouchers  DROP COLUMN IF EXISTS vendor_name;
ALTER TABLE requisitions      DROP COLUMN IF EXISTS preferred_vendor_name;
