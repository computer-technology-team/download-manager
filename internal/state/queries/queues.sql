-- name: CreateQueue :one
INSERT INTO queues (name, directory, max_bandwidth, download_start, download_end, retry_limit)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetQueue :one
SELECT * FROM queues
WHERE id = ?;

-- name: ListQueues :many
SELECT * FROM queues;

-- name: UpdateQueue :one
UPDATE queues
SET name = ?, directory = ?, max_bandwidth = ?, download_start = ?, download_end = ?, retry_limit = ?
WHERE id = ?
RETURNING *;

-- name: DeleteQueue :exec
DELETE FROM queues
WHERE id = ?;
