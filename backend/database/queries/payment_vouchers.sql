-- Payment voucher read queries.
-- Three scope variants:
--   All         → scope.CanViewAll
--   Procurement → scope.IsProcurement (linked_po != '' filter)
--   Limited     → default (owner + workflow_tasks involvement filter)

-- name: CountPaymentVouchersAll :one
SELECT COUNT(*) FROM payment_vouchers
WHERE organization_id = $1
  AND ($2::text = '' OR UPPER(status) = UPPER($2))
  AND ($3::text = '' OR vendor_id     = $3)
  AND (NOT $4::bool OR routing_type != 'direct_payment');

-- name: ListPaymentVoucherIDsAll :many
SELECT id FROM payment_vouchers
WHERE organization_id = $1
  AND ($2::text = '' OR UPPER(status) = UPPER($2))
  AND ($3::text = '' OR vendor_id     = $3)
  AND (NOT $4::bool OR routing_type != 'direct_payment')
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountPaymentVouchersProcurement :one
SELECT COUNT(*) FROM payment_vouchers pv
WHERE pv.organization_id = $1
  AND pv.linked_po IS NOT NULL
  AND pv.linked_po != ''
  AND ($2::text = '' OR UPPER(pv.status) = UPPER($2))
  AND ($3::text = '' OR pv.vendor_id     = $3)
  AND (NOT $4::bool OR pv.routing_type != 'direct_payment');

-- name: ListPaymentVoucherIDsProcurement :many
SELECT pv.id FROM payment_vouchers pv
WHERE pv.organization_id = $1
  AND pv.linked_po IS NOT NULL
  AND pv.linked_po != ''
  AND ($2::text = '' OR UPPER(pv.status) = UPPER($2))
  AND ($3::text = '' OR pv.vendor_id     = $3)
  AND (NOT $4::bool OR pv.routing_type != 'direct_payment')
ORDER BY pv.created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountPaymentVouchersLimited :one
SELECT COUNT(*) FROM payment_vouchers pv
WHERE pv.organization_id = $1
  AND ($2::text = '' OR UPPER(pv.status) = UPPER($2))
  AND ($3::text = '' OR pv.vendor_id     = $3)
  AND (
      pv.created_by = $4
      OR pv.id IN (
          SELECT wt.entity_id FROM workflow_tasks wt
          WHERE wt.organization_id = $1
            AND wt.entity_type     = 'payment_voucher'
            AND (
                wt.assigned_user_id        = $4
                OR LOWER(wt.assigned_role) = LOWER($5)
                OR wt.assigned_role        = ANY($6::text[])
                OR wt.claimed_by           = $4
            )
      )
  );

-- name: ListPaymentVoucherIDsLimited :many
SELECT pv.id FROM payment_vouchers pv
WHERE pv.organization_id = $1
  AND ($2::text = '' OR UPPER(pv.status) = UPPER($2))
  AND ($3::text = '' OR pv.vendor_id     = $3)
  AND (
      pv.created_by = $4
      OR pv.id IN (
          SELECT wt.entity_id FROM workflow_tasks wt
          WHERE wt.organization_id = $1
            AND wt.entity_type     = 'payment_voucher'
            AND (
                wt.assigned_user_id        = $4
                OR LOWER(wt.assigned_role) = LOWER($5)
                OR wt.assigned_role        = ANY($6::text[])
                OR wt.claimed_by           = $4
            )
      )
  )
ORDER BY pv.created_at DESC
LIMIT $7 OFFSET $8;
