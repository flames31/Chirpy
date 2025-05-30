-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid (),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserCredential :exec
UPDATE users
SET email = $2,
hashed_password = $3
WHERE id = $1;

-- name: UpdateChirpyRed :exec
UPDATE users
SET is_chirpy_red = $2
WHERE id = $1;