CREATE TABLE queues (
    ID INTEGER PRIMARY KEY AUTOINCREMENT, -- Auto-incrementing primary key
    Name TEXT NOT NULL, -- Name of the queue
    Directory TEXT NOT NULL, -- Directory to save downloads
    MaxBandwidth INTEGER, -- Max bandwidth in KB/s
    DownloadStart TEXT, -- Start of download window (stored as text)
    DownloadEnd TEXT, -- End of download window (stored as text)
    RetryLimit INTEGER NOT NULL DEFAULT 3 -- Max retry attempts, default is 3
);