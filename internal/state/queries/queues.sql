-- name: CreateQueue :one
INSERT INTO queues (name, directory, max_bandwidth, start_download, end_download, retry_limit)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetQueue :one
SELECT * FROM queues
WHERE id = ?;

-- name: ListQueues :many
SELECT * FROM queues;

-- name: UpdateQueue :one
UPDATE queues
SET name = ?, max_bandwidth = ?, start_download = ?, end_download = ?, retry_limit = ?
WHERE id = ?
RETURNING *;

-- name: DeleteQueue :exec
DELETE FROM queues
WHERE id = ?;
