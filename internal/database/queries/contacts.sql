-- name: CreateContact :one
INSERT INTO contacts (name, email, subject, message, ip_address)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CountContactsByIPToday :one
SELECT COUNT(*) FROM contacts
WHERE ip_address = $1
AND created_at >= CURRENT_DATE
AND created_at < CURRENT_DATE + INTERVAL '1 day';

-- name: CountContactsByEmailToday :one
SELECT COUNT(*) FROM contacts
WHERE email = $1
AND created_at >= CURRENT_DATE
AND created_at < CURRENT_DATE + INTERVAL '1 day';