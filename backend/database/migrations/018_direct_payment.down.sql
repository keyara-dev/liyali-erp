DROP INDEX IF EXISTS idx_payment_vouchers_routing_type_org;
DROP INDEX IF EXISTS idx_purchase_orders_routing_type_org;
DROP INDEX IF EXISTS idx_requisitions_routing_type_org;

ALTER TABLE payment_vouchers
    DROP COLUMN IF EXISTS paid_by,
    DROP COLUMN IF EXISTS paid_at,
    DROP COLUMN IF EXISTS proof_of_payment,
    DROP COLUMN IF EXISTS routing_type;

ALTER TABLE purchase_orders
    DROP COLUMN IF EXISTS routing_type;

ALTER TABLE requisitions
    DROP COLUMN IF EXISTS payee_snapshot,
    DROP COLUMN IF EXISTS payee_id,
    DROP COLUMN IF EXISTS routing_type;

DROP INDEX IF EXISTS idx_payees_name_trgm;
DROP INDEX IF EXISTS idx_payees_org_type_name;
DROP TABLE IF EXISTS payees;
