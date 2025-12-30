-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, name, role, department, is_active, email_verified
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
    name = COALESCE($2, name),
    role = COALESCE($3, role),
    department = COALESCE($4, department),
    is_active = COALESCE($5, is_active),
    email_verified = COALESCE($6, email_verified)
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: IncrementFailedLoginAttempts :exec
UPDATE users
SET failed_login_attempts = failed_login_attempts + 1
WHERE id = $1;

-- name: ResetFailedLoginAttempts :exec
UPDATE users
SET failed_login_attempts = 0, locked_until = NULL
WHERE id = $1;

-- name: LockUserAccount :exec
UPDATE users
SET locked_until = $2
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users
SET is_active = false
WHERE id = $1;

-- name: ActivateUser :exec
UPDATE users
SET is_active = true
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListUsersByRole :many
SELECT * FROM users
WHERE role = $1
ORDER BY created_at DESC;

-- name: ListUsersByDepartment :many
SELECT * FROM users
WHERE department = $1
ORDER BY created_at DESC;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountActiveUsers :one
SELECT COUNT(*) FROM users WHERE is_active = true;
