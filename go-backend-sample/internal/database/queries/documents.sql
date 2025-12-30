-- name: CreateDocument :one
INSERT INTO documents (
    document_type,
    document_number,
    title,
    description,
    amount,
    currency,
    status,
    created_by,
    department,
    workflow_id,
    data,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetDocumentByID :one
SELECT * FROM documents
WHERE id = $1;

-- name: GetDocumentByNumber :one
SELECT * FROM documents
WHERE document_number = $1;

-- name: ListDocuments :many
SELECT * FROM documents
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListDocumentsByType :many
SELECT * FROM documents
WHERE document_type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDocumentsByStatus :many
SELECT * FROM documents
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDocumentsByCreator :many
SELECT * FROM documents
WHERE created_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDocumentsByDepartment :many
SELECT * FROM documents
WHERE department = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDocumentsByTypeAndStatus :many
SELECT * FROM documents
WHERE document_type = $1 AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListDocumentsByWorkflow :many
SELECT * FROM documents
WHERE workflow_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateDocument :one
UPDATE documents
SET title = $2,
    description = $3,
    amount = $4,
    currency = $5,
    data = $6,
    metadata = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDocumentStatus :one
UPDATE documents
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SubmitDocument :one
UPDATE documents
SET status = 'SUBMITTED',
    submitted_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ApproveDocument :one
UPDATE documents
SET status = 'APPROVED',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RejectDocument :one
UPDATE documents
SET status = 'REJECTED',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDocument :exec
DELETE FROM documents
WHERE id = $1;

-- name: CountDocuments :one
SELECT COUNT(*) FROM documents;

-- name: CountDocumentsByType :one
SELECT COUNT(*) FROM documents
WHERE document_type = $1;

-- name: CountDocumentsByStatus :one
SELECT COUNT(*) FROM documents
WHERE status = $1;

-- name: CountDocumentsByCreator :one
SELECT COUNT(*) FROM documents
WHERE created_by = $1;

-- name: GetDocumentAmountSum :one
SELECT COALESCE(SUM(amount), 0) FROM documents
WHERE status = $1 AND created_at >= $2;
