-- name: CreateWorkflow :one
INSERT INTO workflows (
    name,
    description,
    document_type,
    stages,
    is_active,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetWorkflowByID :one
SELECT * FROM workflows
WHERE id = $1;

-- name: ListWorkflows :many
SELECT * FROM workflows
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListActiveWorkflows :many
SELECT * FROM workflows
WHERE is_active = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListWorkflowsByDocumentType :many
SELECT * FROM workflows
WHERE document_type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveWorkflowsByDocumentType :many
SELECT * FROM workflows
WHERE document_type = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetDefaultWorkflowByDocumentType :one
SELECT * FROM workflows
WHERE document_type = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateWorkflow :one
UPDATE workflows
SET name = $2,
    description = $3,
    stages = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ActivateWorkflow :one
UPDATE workflows
SET is_active = true,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeactivateWorkflow :one
UPDATE workflows
SET is_active = false,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteWorkflow :exec
DELETE FROM workflows
WHERE id = $1;

-- name: CountWorkflows :one
SELECT COUNT(*) FROM workflows;

-- name: CountActiveWorkflows :one
SELECT COUNT(*) FROM workflows
WHERE is_active = true;

-- name: CountWorkflowsByDocumentType :one
SELECT COUNT(*) FROM workflows
WHERE document_type = $1;
