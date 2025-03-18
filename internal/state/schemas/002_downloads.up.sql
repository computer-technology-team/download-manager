CREATE TABLE downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Auto-incrementing primary key
    queue_id INTEGER NOT NULL, -- Foreign key to the Queue table
    url TEXT NOT NULL,
    save_path TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'PAUSED',
    retries INTEGER DEFAULT 0,

    FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);

CREATE TABLE download_chunks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    range_start INTEGER NOT NULL,
    range_end INTEGER NOT NULL,
    current_pointer INTEGER NOT NULL,
    download_id INTEGER NOT NULL,

    FOREIGN KEY (download_id) REFERENCES downloads(id) ON DELETE CASCADE
);
