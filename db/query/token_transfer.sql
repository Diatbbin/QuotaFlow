-- name: CreateTokenTransfer :one
INSERT INTO token_transfers (
    from_ai_tool_id,
    to_ai_tool_id,
    tokens
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetTokenTransfer :one
SELECT * FROM token_transfers
WHERE id = $1 LIMIT 1;

-- name: ListTokenTransfers :many
SELECT * FROM token_transfers
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: UpdateTokenTransferTokens :one
UPDATE token_transfers
SET tokens = $1
WHERE id = $2
RETURNING *;

-- name: DeleteTokenTransfer :exec
DELETE FROM token_transfers
WHERE id = $1;
