CREATE TABLE downloads (
    ID INTEGER PRIMARY KEY AUTOINCREMENT, -- Auto-incrementing primary key
    QueueID INTEGER NOT NULL -- Foreign key to the Queue table
);

-- name: GetDownloadsByQueueID :many
SELECT * FROM downloads
WHERE QueueID = ?;

-- name: CountDownloadsInQueue :one
SELECT COUNT(*) FROM downloads
WHERE QueueID = ?;
