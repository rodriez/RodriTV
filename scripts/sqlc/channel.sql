-- name: ListChannels :many
SELECT * FROM channels
ORDER BY number;

-- name: GetChannelByID :one
SELECT * FROM channels
WHERE id = $1;

-- name: CreateChannel :one
INSERT INTO channels (
    id, number, name, logo, schedule_url, enabled, segment_seconds,
    sources, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateChannel :one
UPDATE channels
SET number = $2,
    name = $3,
    logo = $4,
    schedule_url = $5,
    enabled = $6,
    segment_seconds = $7,
    sources = $8,
    updated_at = $9
WHERE id = $1
RETURNING *;

-- name: DeleteChannel :execrows
DELETE FROM channels
WHERE id = $1;

-- name: ExistsChannelNumber :one
SELECT EXISTS(
    SELECT 1 FROM channels WHERE number = $1 AND id <> $2
);
