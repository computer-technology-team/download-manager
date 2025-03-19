-- name: CreateDownload :one
INSERT INTO downloads (queue_id, url, save_path, state, retries)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetDownload :one
SELECT * FROM downloads
WHERE id = ?;

-- name: ListDownloads :many
SELECT * FROM downloads;

-- name: UpdateDownload :one
UPDATE downloads
SET queue_id = ?, url = ?, save_path = ?, state = ?, retries = ?
WHERE id = ?
RETURNING *;

-- name: DeleteDownload :exec
DELETE FROM downloads
WHERE id = ?;

-- name: UpsertDownload :one
INSERT INTO downloads (id, queue_id, url, save_path, state, retries)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE
SET queue_id = EXCLUDED.queue_id,
    url = EXCLUDED.url,
    save_path = EXCLUDED.save_path,
    state = EXCLUDED.state,
    retries = EXCLUDED.retries
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