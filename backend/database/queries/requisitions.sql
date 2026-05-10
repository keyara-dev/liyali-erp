-- Requisition read queries.
-- List queries return only `id`; the handler fetches full data via GORM with Preload
-- to avoid complex pgtype↔model conversions. This keeps mappers unchanged.
-- Three scope variants per list/count pair:
--   All         → scope.CanViewAll
--   Procurement → scope.IsProcurement (workflow_assignments subquery filter)
--   Limited     → default (owner + workflow_tasks involvement filter)

-- name: CountRequisitionsAll :one
SELECT COUNT(*) FROM requisitions
WHERE organization_id = $1
  AND ($2::text = '' OR UPPER(status)     = UPPER($2))
  AND ($3::text = '' OR department        = $3)
  AND ($4::text = '' OR priority          = $4)
  AND (NOT $5::bool OR routing_type != 'direct_payment');

-- name: ListRequisitionIDsAll :many
SELECT id FROM requisitions
WHERE organization_id = $1
  AND ($2::text = '' OR UPPER(status)     = UPPER($2))
  AND ($3::text = '' OR department        = $3)
  AND ($4::text = '' OR priority          = $4)
  AND (NOT $5::bool OR routing_type != 'direct_payment')
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountRequisitionsProcurement :one
SELECT COUNT(*) FROM requisitions r
WHERE r.organization_id = $1
  AND ($2::text = '' OR UPPER(r.status)   = UPPER($2))
  AND ($3::text = '' OR r.department      = $3)
  AND ($4::text = '' OR r.priority        = $4)
  AND (NOT $5::bool OR r.routing_type != 'direct_payment')
  AND r.id IN (
      SELECT wa.entity_id
      FROM workflow_assignments wa
      JOIN workflows w ON w.id = wa.workflow_id
      WHERE wa.entity_type     = 'requisition'
        AND wa.organization_id = $1
        AND (
            w.conditions->>'routingType' IS NULL OR
            w.conditions->>'routingType' = ''   OR
            w.conditions->>'routingType' = 'procurement'
        )
  );

-- name: ListRequisitionIDsProcurement :many
SELECT r.id FROM requisitions r
WHERE r.organization_id = $1
  AND ($2::text = '' OR UPPER(r.status)   = UPPER($2))
  AND ($3::text = '' OR r.department      = $3)
  AND ($4::text = '' OR r.priority        = $4)
  AND (NOT $5::bool OR r.routing_type != 'direct_payment')
  AND r.id IN (
      SELECT wa.entity_id
      FROM workflow_assignments wa
      JOIN workflows w ON w.id = wa.workflow_id
      WHERE wa.entity_type     = 'requisition'
        AND wa.organization_id = $1
        AND (
            w.conditions->>'routingType' IS NULL OR
            w.conditions->>'routingType' = ''   OR
            w.conditions->>'routingType' = 'procurement'
        )
  )
ORDER BY r.created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountRequisitionsLimited :one
SELECT COUNT(*) FROM requisitions r
WHERE r.organization_id = $1
  AND ($2::text = '' OR UPPER(r.status)     = UPPER($2))
  AND ($3::text = '' OR r.department        = $3)
  AND ($4::text = '' OR r.priority          = $4)
  AND (
      r.requester_id = $5
      OR r.id IN (
          SELECT wt.entity_id FROM workflow_tasks wt
          WHERE wt.organization_id = $1
            AND wt.entity_type     = 'requisition'
            AND (
                wt.assigned_user_id        = $5
                OR LOWER(wt.assigned_role) = LOWER($6)
                OR wt.assigned_role        = ANY($7::text[])
                OR wt.claimed_by           = $5
            )
      )
  );

-- name: ListRequisitionIDsLimited :many
SELECT r.id FROM requisitions r
WHERE r.organization_id = $1
  AND ($2::text = '' OR UPPER(r.status)     = UPPER($2))
  AND ($3::text = '' OR r.department        = $3)
  AND ($4::text = '' OR r.priority          = $4)
  AND (
      r.requester_id = $5
      OR r.id IN (
          SELECT wt.entity_id FROM workflow_tasks wt
          WHERE wt.organization_id = $1
            AND wt.entity_type     = 'requisition'
            AND (
                wt.assigned_user_id        = $5
                OR LOWER(wt.assigned_role) = LOWER($6)
                OR wt.assigned_role        = ANY($7::text[])
                OR wt.claimed_by           = $5
            )
      )
  )
ORDER BY r.created_at DESC
LIMIT $8 OFFSET $9;
