-- Make vendor_id nullable in purchase_orders table to allow PO creation without vendor
-- Migration: 004_make_vendor_id_nullable
-- Description: Allows creating purchase orders without specifying a vendor
-- Date: 2025-01-09

-- Remove the NOT NULL constraint from vendor_id
ALTER TABLE purchase_orders ALTER COLUMN vendor_id DROP NOT NULL;

-- Remove the foreign key constraint temporarily
ALTER TABLE purchase_orders DROP CONSTRAINT IF EXISTS fk_purchase_orders_vendor;

-- Add the foreign key constraint back but allow NULL values
ALTER TABLE purchase_orders 
ADD CONSTRAINT fk_purchase_orders_vendor 
FOREIGN KEY (vendor_id) REFERENCES vendors(id) 
ON DELETE SET NULL;

-- Add a comment to document the change
COMMENT ON COLUMN purchase_orders.vendor_id IS 'Vendor ID - nullable to allow PO creation without vendor';

SELECT 'Vendor ID column made nullable successfully' as status;