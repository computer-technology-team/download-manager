-- name: CreateQueue :one
INSERT INTO queues (Name, Directory, MaxBandwidth, DownloadStart, DownloadEnd, RetryLimit)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetQueue :one
SELECT * FROM queues
WHERE ID = ?;

-- name: ListQueues :many
SELECT * FROM queues;

-- name: UpdateQueue :one
UPDATE queues
SET Name = ?, Directory = ?, MaxBandwidth = ?, DownloadStart = ?, DownloadEnd = ?, RetryLimit = ?
WHERE ID = ?
RETURNING *;

-- name: DeleteQueue :exec
DELETE FROM queues
WHERE ID = ?;