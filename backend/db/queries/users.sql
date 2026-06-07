-- name: CreateUser :one
INSERT INTO users (email, hashed_password, role, first_name, last_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET first_name = $2,
    last_name  = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetUserVerified :exec
UPDATE users
SET is_verified = TRUE,
    updated_at  = now()
WHERE id = $1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountUsers :one
SELECT count(*) FROM users;
