-- name: CreateDownload :one
INSERT INTO downloads (queue_id, url, save_path, state, retries)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetDownload :one
SELECT * FROM downloads
WHERE id = ?;

-- name: ListDownloadsWithQueueName :many
SELECT downloads.*, queues.name as queue_name
FROM downloads JOIN queues on downloads.queue_id = queues.id;

-- name: ListDownloads :many
SELECT * 
FROM downloads;

-- name: SetDownloadState :one
UPDATE downloads
SET state = ?
WHERE id = ?
RETURNING *;

-- name: SetDownloadRetry :one
UPDATE downloads
SET retries = ?
WHERE id = ?
RETURNING *;

-- name: DeleteDownload :exec
DELETE FROM downloads
WHERE id = ?;

-- name: UpsertDownloadChunk :one
INSERT INTO download_chunks (id, range_start, range_end, current_pointer, download_id)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE
SET current_pointer = EXCLUDED.current_pointer
RETURNING *;

-- name: GetDownloadChunk :one
SELECT * FROM download_chunks
WHERE id = ?;

-- name: ListDownloadChunks :many
SELECT * FROM download_chunks;

-- name: DeleteDownloadChunk :exec
DELETE FROM download_chunks
WHERE id = ?;

-- name: GetDownloadChunksByDownloadID :many
SELECT * FROM download_chunks
WHERE download_id = ?;

-- name: GetPendingDownloadByQueueID :one
SELECT * FROM downloads
WHERE queue_id = ? AND state = 'PENDING'
LIMIT 1;
