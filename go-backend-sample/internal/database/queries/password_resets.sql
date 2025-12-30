-- name: CreatePasswordReset :one
INSERT INTO password_resets (
    user_id, token, expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetPasswordResetByToken :one
SELECT * FROM password_resets
WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP AND used = false
LIMIT 1;

-- name: MarkPasswordResetAsUsed :exec
UPDATE password_resets
SET used = true
WHERE id = $1;

-- name: DeletePasswordResetsByUserID :exec
DELETE FROM password_resets WHERE user_id = $1;

-- name: DeleteExpiredPasswordResets :exec
DELETE FROM password_resets WHERE expires_at <= CURRENT_TIMESTAMP;

-- name: DeleteUsedPasswordResets :exec
DELETE FROM password_resets WHERE used = true;
