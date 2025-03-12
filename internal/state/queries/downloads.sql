-- name: GetAllDownloads :many
SELECT * FROM downloads;

-- name: CreateDownload :exec
INSERT INTO downloads (url, save_path, bandwidth_limit_bytes_p_s)
VALUES (?, ?, ?);

-- name: UpdateDownloadState :exec
UPDATE downloads
SET state = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetDownloadStatus :one
SELECT state, progress_persent, speed_bytes_p_s FROM downloads
WHERE id = ?;
