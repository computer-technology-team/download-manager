CREATE TABLE queues (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Auto-incrementing primary key
    name TEXT NOT NULL, -- Name of the queue
    directory TEXT NOT NULL, -- Directory to save downloads
    max_bandwidth INTEGER, -- Max bandwidth in KB/s
    start_download TEXT NOT NULL, -- Start of download window (stored as text)
    end_download TEXT NOT NULL, -- End of download window (stored as text)
    retry_limit INTEGER NOT NULL DEFAULT 3 -- Max retry attempts, default is 3
);
