-- name: CreateApprovalTask :one
INSERT INTO approval_tasks (
    document_id,
    assigned_to,
    assigned_by,
    status,
    current_stage,
    total_stages,
    priority,
    due_date,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetApprovalTaskByID :one
SELECT * FROM approval_tasks
WHERE id = $1;

-- name: ListApprovalTasksByAssignee :many
SELECT * FROM approval_tasks
WHERE assigned_to = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListApprovalTasksByStatus :many
SELECT * FROM approval_tasks
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListApprovalTasksByAssigneeAndStatus :many
SELECT * FROM approval_tasks
WHERE assigned_to = $1 AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListApprovalTasksByDocument :many
SELECT * FROM approval_tasks
WHERE document_id = $1
ORDER BY current_stage ASC, created_at DESC;

-- name: ListPendingApprovalTasks :many
SELECT * FROM approval_tasks
WHERE status = 'PENDING'
ORDER BY priority DESC, created_at ASC
LIMIT $1 OFFSET $2;

-- name: ListOverdueApprovalTasks :many
SELECT * FROM approval_tasks
WHERE status = 'PENDING'
  AND due_date < NOW()
ORDER BY due_date ASC
LIMIT $1 OFFSET $2;

-- name: UpdateApprovalTaskStatus :one
UPDATE approval_tasks
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateApprovalTaskStage :one
UPDATE approval_tasks
SET current_stage = $2,
    status = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ReassignApprovalTask :one
UPDATE approval_tasks
SET assigned_to = $2,
    assigned_by = $3,
    status = 'REASSIGNED',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateApprovalTaskNotes :one
UPDATE approval_tasks
SET notes = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteApprovalTask :exec
DELETE FROM approval_tasks
WHERE id = $1;

-- name: CountApprovalTasksByAssignee :one
SELECT COUNT(*) FROM approval_tasks
WHERE assigned_to = $1;

-- name: CountApprovalTasksByStatus :one
SELECT COUNT(*) FROM approval_tasks
WHERE status = $1;

-- name: CountPendingApprovalTasksByAssignee :one
SELECT COUNT(*) FROM approval_tasks
WHERE assigned_to = $1 AND status = 'PENDING';
