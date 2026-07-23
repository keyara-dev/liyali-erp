-- Enhanced user queries with security features

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (
    id, email, name, password, role, active, current_organization_id, is_super_admin
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET 
    name = COALESCE($2, name),
    email = COALESCE($3, email),
    role = COALESCE($4, role),
    active = COALESCE($5, active),
    current_organization_id = COALESCE($6, current_organization_id),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET 
    password = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users SET 
    last_login = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users SET 
    active = false,
    updated_at = NOW()
WHERE id = $1;

-- name: ActivateUser :exec
UPDATE users SET 
    active = true,
    updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users 
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListUsersByOrganization :many
SELECT u.* FROM users u
INNER JOIN organization_members om ON u.id = om.user_id
WHERE om.organization_id = $1 AND om.active = true
ORDER BY u.name
LIMIT $2 OFFSET $3;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountActiveUsers :one
SELECT COUNT(*) FROM users WHERE active = true;