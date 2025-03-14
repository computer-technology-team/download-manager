CREATE TABLE queues (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Auto-incrementing primary key
    name TEXT NOT NULL, -- Name of the queue
    directory TEXT NOT NULL, -- Directory to save downloads
    max_bandwidth INTEGER, -- Max bandwidth in KB/s
    download_start TEXT, -- Start of download window (stored as text)
    download_end TEXT, -- End of download window (stored as text)
    retry_limit INTEGER NOT NULL DEFAULT 3 -- Max retry attempts, default is 3
);
