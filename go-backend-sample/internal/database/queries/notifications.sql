-- name: CreateNotification :one
INSERT INTO notifications (
    user_id,
    type,
    title,
    message,
    related_id,
    sent_via_email
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetNotificationByID :one
SELECT * FROM notifications
WHERE id = $1;

-- name: ListNotificationsByUser :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUnreadNotificationsByUser :many
SELECT * FROM notifications
WHERE user_id = $1 AND is_read = false
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListNotificationsByType :many
SELECT * FROM notifications
WHERE type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListNotificationsByRelatedID :many
SELECT * FROM notifications
WHERE related_id = $1
ORDER BY created_at DESC;

-- name: MarkNotificationAsRead :one
UPDATE notifications
SET is_read = true
WHERE id = $1
RETURNING *;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET is_read = true
WHERE user_id = $1 AND is_read = false;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1;

-- name: DeleteOldNotifications :exec
DELETE FROM notifications
WHERE created_at < $1;

-- name: CountUnreadNotificationsByUser :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1 AND is_read = false;

-- name: CountNotificationsByUser :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1;

-- name: CountNotificationsByType :one
SELECT COUNT(*) FROM notifications
WHERE type = $1;
