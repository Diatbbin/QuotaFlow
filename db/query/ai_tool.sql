-- name: CreateAiTool :one
INSERT INTO ai_tools (
    username,
    tool,
    token_limit
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetAiTool :one
SELECT * FROM ai_tools
WHERE id = $1 LIMIT 1;

-- name: GetAiToolForUpdate :one
SELECT * FROM ai_tools
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListAiTools :many
SELECT * FROM ai_tools
WHERE username = $1
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: AddAiToolTokenLimit :one
UPDATE ai_tools
SET token_limit = token_limit + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateAiToolTokenLimit :one
UPDATE ai_tools
SET token_limit = $1
WHERE id = $2
RETURNING *;

-- name: DeleteAiTool :exec
DELETE FROM ai_tools
WHERE id = $1;
