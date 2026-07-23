-- Purchase order read queries.
-- Both CanViewAll and IsProcurement return all POs (ApplyToQuery passes through for both).
-- Only two scope variants needed: All and Limited.

-- name: CountPurchaseOrdersAll :one
SELECT COUNT(*) FROM purchase_orders
WHERE organization_id = $1
  AND ($2::text = '' OR UPPER(status)    = UPPER($2))
  AND ($3::text = '' OR vendor_id        = $3)
  AND (NOT $4::bool OR routing_type != 'direct_payment');

-- name: ListPurchaseOrderIDsAll :many
SELECT id FROM purchase_orders
WHERE organization_id = $1
  AND ($2::text = '' OR UPPER(status)    = UPPER($2))
  AND ($3::text = '' OR vendor_id        = $3)
  AND (NOT $4::bool OR routing_type != 'direct_payment')
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountPurchaseOrdersLimited :one
SELECT COUNT(*) FROM purchase_orders po
WHERE po.organization_id = $1
  AND ($2::text = '' OR UPPER(po.status) = UPPER($2))
  AND ($3::text = '' OR po.vendor_id     = $3)
  AND (
      po.created_by = $4
      OR po.id IN (
          SELECT wt.entity_id FROM workflow_tasks wt
          WHERE wt.organization_id = $1
            AND wt.entity_type     = 'purchase_order'
            AND (
                wt.assigned_user_id        = $4
                OR LOWER(wt.assigned_role) = LOWER($5)
                OR wt.assigned_role        = ANY($6::text[])
                OR wt.claimed_by           = $4
            )
      )
  );

-- name: ListPurchaseOrderIDsLimited :many
SELECT po.id FROM purchase_orders po
WHERE po.organization_id = $1
  AND ($2::text = '' OR UPPER(po.status) = UPPER($2))
  AND ($3::text = '' OR po.vendor_id     = $3)
  AND (
      po.created_by = $4
      OR po.id IN (
          SELECT wt.entity_id FROM workflow_tasks wt
          WHERE wt.organization_id = $1
            AND wt.entity_type     = 'purchase_order'
            AND (
                wt.assigned_user_id        = $4
                OR LOWER(wt.assigned_role) = LOWER($5)
                OR wt.assigned_role        = ANY($6::text[])
                OR wt.claimed_by           = $4
            )
      )
  )
ORDER BY po.created_at DESC
LIMIT $7 OFFSET $8;
