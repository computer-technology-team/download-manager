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

-- name: CreateDownloadChunk :one
INSERT INTO download_chunks (id, range_start, range_end, current_pointer, download_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetDownloadChunk :one
SELECT * FROM download_chunks
WHERE id = ?;

-- name: ListDownloadChunks :many
SELECT * FROM download_chunks;

-- name: UpdateDownloadChunk :one
UPDATE download_chunks
SET range_start = ?, range_end = ?, current_pointer = ?, download_id = ?
WHERE id = ?
RETURNING *;

-- name: DeleteDownloadChunk :exec
DELETE FROM download_chunks
WHERE id = ?;