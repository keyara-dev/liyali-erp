-- Rollback vendor_id nullable change in purchase_orders table
-- Migration: 004_make_vendor_id_nullable (DOWN)
-- Description: Makes vendor_id NOT NULL again and restores original constraint
-- Date: 2025-01-10

-- First, we need to handle any NULL vendor_id values
-- Update NULL vendor_id values to a default vendor or remove those records
-- For safety, we'll just report if there are any NULL values
DO $
DECLARE
    null_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO null_count FROM purchase_orders WHERE vendor_id IS NULL;
    
    IF null_count > 0 THEN
        RAISE NOTICE 'WARNING: Found % purchase orders with NULL vendor_id', null_count;
        RAISE NOTICE 'These records need to be handled before making vendor_id NOT NULL';
        RAISE EXCEPTION 'Cannot rollback: NULL vendor_id values exist';
    END IF;
END $;

-- Remove the current foreign key constraint
ALTER TABLE purchase_orders DROP CONSTRAINT IF EXISTS fk_purchase_orders_vendor;

-- Make vendor_id NOT NULL again
ALTER TABLE purchase_orders ALTER COLUMN vendor_id SET NOT NULL;

-- Add the foreign key constraint back without NULL handling
ALTER TABLE purchase_orders 
ADD CONSTRAINT fk_purchase_orders_vendor 
FOREIGN KEY (vendor_id) REFERENCES vendors(id);

-- Remove the comment about nullable vendor_id
COMMENT ON COLUMN purchase_orders.vendor_id IS 'Vendor ID - required for purchase orders';

SELECT 'Vendor ID nullable rollback completed successfully' as status;