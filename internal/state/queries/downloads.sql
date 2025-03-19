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

-- name: UpsertDownloadChunk :one
INSERT INTO download_chunks (id, range_start, range_end, current_pointer, download_id)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE
SET range_start = EXCLUDED.range_start,
    range_end = EXCLUDED.range_end,
    current_pointer = EXCLUDED.current_pointer,
    download_id = EXCLUDED.download_id
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