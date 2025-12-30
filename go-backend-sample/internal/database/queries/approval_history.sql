-- name: CreateApprovalHistoryEntry :one
INSERT INTO approval_history (
    task_id,
    user_id,
    action,
    stage,
    comment,
    signature,
    ip_address
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetApprovalHistoryByID :one
SELECT * FROM approval_history
WHERE id = $1;

-- name: ListApprovalHistoryByTask :many
SELECT * FROM approval_history
WHERE task_id = $1
ORDER BY created_at ASC;

-- name: ListApprovalHistoryByUser :many
SELECT * FROM approval_history
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListApprovalHistoryByAction :many
SELECT * FROM approval_history
WHERE action = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetLatestApprovalHistoryByTask :one
SELECT * FROM approval_history
WHERE task_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteApprovalHistory :exec
DELETE FROM approval_history
WHERE id = $1;

-- name: CountApprovalHistoryByTask :one
SELECT COUNT(*) FROM approval_history
WHERE task_id = $1;

-- name: CountApprovalHistoryByUser :one
SELECT COUNT(*) FROM approval_history
WHERE user_id = $1;
