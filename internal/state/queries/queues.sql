-- name: CreateQueue :one
INSERT INTO queues (name, directory, max_bandwidth, start_download, end_download, retry_limit, max_concurrent, schedule_mode)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetQueue :one
SELECT * FROM queues
WHERE id = ?;

-- name: ListQueues :many
SELECT * FROM queues;

-- name: UpdateQueue :one
UPDATE queues
SET name = ?, max_bandwidth = ?, start_download = ?, end_download = ?,
retry_limit = ?, max_concurrent = ?, schedule_mode = ?, directory = ?
WHERE id = ?
RETURNING *;

-- name: DeleteQueue :exec
DELETE FROM queues
WHERE id = ?;

-- name: CountInProgressDownloadsInQueue :one
SELECT COUNT(*) 
FROM downloads
WHERE queue_id = ? AND status = 'IN_PROGRESS';

-- name: UpdateInProgressToPendingInQueue :exec
UPDATE downloads
SET status = 'PENDING'
WHERE status = 'IN_PROGRESS' AND queue_id = ?;