-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    user_id,
    action,
    resource_type,
    resource_id,
    changes,
    ip_address,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetAuditLogByID :one
SELECT * FROM audit_logs
WHERE id = $1;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByResource :many
SELECT * FROM audit_logs
WHERE resource_type = $1 AND resource_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListAuditLogsByResourceType :many
SELECT * FROM audit_logs
WHERE resource_type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByAction :many
SELECT * FROM audit_logs
WHERE action = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByDateRange :many
SELECT * FROM audit_logs
WHERE created_at BETWEEN $1 AND $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: DeleteOldAuditLogs :exec
DELETE FROM audit_logs
WHERE created_at < $1;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs;

-- name: CountAuditLogsByUser :one
SELECT COUNT(*) FROM audit_logs
WHERE user_id = $1;

-- name: CountAuditLogsByResource :one
SELECT COUNT(*) FROM audit_logs
WHERE resource_type = $1 AND resource_id = $2;
