-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT * FROM users WHERE id IN (SELECT user_id FROM refresh_tokens WHERE token = $1);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(),
updated_at = NOW()
WHERE token = $1;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1;