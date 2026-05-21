-- name: CreateWorkspace :one
INSERT INTO workspaces (
    name,
    token_limit
) VALUES (
    $1, $2
)
RETURNING *;

-- name: GetWorkspace :one
SELECT * FROM workspaces
WHERE id = $1 LIMIT 1;

-- name: GetWorkspaceForUpdate :one
SELECT * FROM workspaces
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListWorkspaces :many
SELECT * FROM workspaces
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: AddWorkspaceTokenLimit :one
UPDATE workspaces
SET token_limit = token_limit + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces
WHERE id = $1;
