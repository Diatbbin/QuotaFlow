-- name: CreateUser :one
INSERT INTO users (
    email,
    username,
    password_hash
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;
