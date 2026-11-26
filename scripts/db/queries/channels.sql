-- name: GetChannelByID :one
SELECT  id, name, list_position, status, signals, created_at, updated_at
FROM channels
WHERE id = $1;

-- name: ListChannels :many
SELECT id, name, list_position, status, signals, created_at, updated_at
FROM channels
ORDER BY list_position;

-- name: CreateChannel :one
INSERT INTO channels (id, name, list_position, status, signals, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, name, list_position, status, signals, created_at, updated_at;

-- name: ReplaceChannel :execrows
UPDATE channels
SET name = $2, list_position = $3, status = $4, signals = $5, updated_at = $6
WHERE id = $1;

-- name: DeleteChannel :exec
DELETE FROM channels
WHERE id = $1;