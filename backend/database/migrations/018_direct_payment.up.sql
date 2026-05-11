-- ============================================================================
-- DIRECT PAYMENT FLOW
-- Migration: 018_direct_payment
-- Adds payees table, routing_type denormalized column on requisitions/POs/PVs,
-- and proof-of-payment fields on payment_vouchers.
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS payees (
    id               VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organization_id  VARCHAR(255) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    payee_type       TEXT NOT NULL CHECK (payee_type IN ('vendor','employee','other')),
    name             TEXT NOT NULL,
    email            TEXT,
    phone            TEXT,
    bank_name        TEXT,
    bank_account     TEXT,
    tax_id           TEXT,
    source_vendor_id VARCHAR(255) NULL REFERENCES vendors(id) ON DELETE SET NULL,
    source_user_id   VARCHAR(255) NULL REFERENCES users(id) ON DELETE SET NULL,
    deleted_at       TIMESTAMPTZ NULL,
    created_by       VARCHAR(255) REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payees_org_type_name ON payees (organization_id, payee_type, name);
CREATE INDEX IF NOT EXISTS idx_payees_name_trgm    ON payees USING gin (name gin_trgm_ops);

ALTER TABLE requisitions
    ADD COLUMN IF NOT EXISTS routing_type   TEXT NOT NULL DEFAULT 'procurement',
    ADD COLUMN IF NOT EXISTS payee_id       VARCHAR(255) REFERENCES payees(id),
    ADD COLUMN IF NOT EXISTS payee_snapshot JSONB;

ALTER TABLE purchase_orders
    ADD COLUMN IF NOT EXISTS routing_type TEXT NOT NULL DEFAULT 'procurement';

ALTER TABLE payment_vouchers
    ADD COLUMN IF NOT EXISTS routing_type     TEXT NOT NULL DEFAULT 'procurement',
    ADD COLUMN IF NOT EXISTS proof_of_payment JSONB,
    ADD COLUMN IF NOT EXISTS paid_at          TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS paid_by          VARCHAR(255) REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_requisitions_routing_type_org
    ON requisitions (organization_id, routing_type);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_routing_type_org
    ON purchase_orders (organization_id, routing_type);
CREATE INDEX IF NOT EXISTS idx_payment_vouchers_routing_type_org
    ON payment_vouchers (organization_id, routing_type);

-- Backfill from workflow conditions for existing rows.
-- requisitions: join via workflow_assignments (entity_id = requisition.id, entity_type = 'requisition')
UPDATE requisitions r
SET routing_type = COALESCE(NULLIF(w.conditions->>'routingType', ''), 'procurement')
FROM workflow_assignments wa
JOIN workflows w ON wa.workflow_id = w.id
WHERE wa.entity_id = r.id
  AND wa.entity_type = 'requisition';

-- purchase_orders: inherit routing_type from the linked requisition
UPDATE purchase_orders po
SET routing_type = r.routing_type
FROM requisitions r
WHERE po.source_requisition_id = r.id;

-- payment_vouchers: inherit routing_type from the linked purchase_order via document_number
UPDATE payment_vouchers pv
SET routing_type = po.routing_type
FROM purchase_orders po
WHERE pv.linked_po = po.document_number;
