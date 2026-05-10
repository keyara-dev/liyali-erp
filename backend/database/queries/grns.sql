-- Goods received note read queries.
-- Both CanViewAll and IsProcurement use the unfiltered path (ApplyToQuery passes through).
-- Limited adds owner (created_by OR received_by) + workflow_tasks involvement.

-- name: CountGRNsAll :one
SELECT COUNT(*) FROM goods_received_notes
WHERE organization_id = $1
  AND ($2::text = '' OR UPPER(status)           = UPPER($2))
  AND ($3::text = '' OR po_document_number      = $3)
  AND (NOT $4::bool OR EXISTS (
        SELECT 1 FROM purchase_orders po
        WHERE po.document_number = po_document_number
          AND po.routing_type != 'direct_payment'
      ));

-- name: ListGRNIDsAll :many
SELECT id FROM goods_received_notes
WHERE organization_id = $1
  AND ($2::text = '' OR UPPER(status)           = UPPER($2))
  AND ($3::text = '' OR po_document_number      = $3)
  AND (NOT $4::bool OR EXISTS (
        SELECT 1 FROM purchase_orders po
        WHERE po.document_number = po_document_number
          AND po.routing_type != 'direct_payment'
      ))
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountGRNsLimited :one
SELECT COUNT(*) FROM goods_received_notes g
WHERE g.organization_id = $1
  AND ($2::text = '' OR UPPER(g.status)          = UPPER($2))
  AND ($3::text = '' OR g.po_document_number     = $3)
  AND (
      g.created_by  = $4
      OR g.received_by = $4
      OR g.id IN (
          SELECT wt.entity_id FROM workflow_tasks wt
          WHERE wt.organization_id = $1
            AND wt.entity_type     = 'grn'
            AND (
                wt.assigned_user_id        = $4
                OR LOWER(wt.assigned_role) = LOWER($5)
                OR wt.assigned_role        = ANY($6::text[])
                OR wt.claimed_by           = $4
            )
      )
  );

-- name: ListGRNIDsLimited :many
SELECT g.id FROM goods_received_notes g
WHERE g.organization_id = $1
  AND ($2::text = '' OR UPPER(g.status)          = UPPER($2))
  AND ($3::text = '' OR g.po_document_number     = $3)
  AND (
      g.created_by  = $4
      OR g.received_by = $4
      OR g.id IN (
          SELECT wt.entity_id FROM workflow_tasks wt
          WHERE wt.organization_id = $1
            AND wt.entity_type     = 'grn'
            AND (
                wt.assigned_user_id        = $4
                OR LOWER(wt.assigned_role) = LOWER($5)
                OR wt.assigned_role        = ANY($6::text[])
                OR wt.claimed_by           = $4
            )
      )
  )
ORDER BY g.created_at DESC
LIMIT $7 OFFSET $8;
