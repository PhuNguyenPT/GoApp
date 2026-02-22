-- internal/database/queries/sessions.sql

-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE token = $1 AND expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < NOW();

-- name: GetActiveSessionsByUserID :many
SELECT * FROM sessions
WHERE user_id = $1 AND expires_at > NOW();

-- name: DeleteSessionByID :exec
DELETE FROM sessions
WHERE id = $1 AND user_id = $2;