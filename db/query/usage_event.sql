-- name: CreateUsageEvent :one
INSERT INTO usage_events (
    workspace_id,
    tokens
) VALUES (
    $1, $2
)
RETURNING *;

-- name: GetUsageEvent :one
SELECT * FROM usage_events
WHERE id = $1 LIMIT 1;

-- name: ListUsageEvents :many
SELECT * FROM usage_events
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: UpdateUsageEventTokens :one
UPDATE usage_events
SET tokens = $1
WHERE id = $2
RETURNING *;

-- name: DeleteUsageEvent :exec
DELETE FROM usage_events
WHERE id = $1;
