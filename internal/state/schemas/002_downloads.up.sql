CREATE TABLE downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Auto-incrementing primary key
    queue_id INTEGER NOT NULL, -- Foreign key to the Queue table
    url TEXT NOT NULL,
    save_path TEXT NOT NULL UNIQUE,
    state TEXT NOT NULL DEFAULT 'PAUSED',
    retries INTEGER NOT NULL DEFAULT 0,

    FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);

CREATE TABLE download_chunks (
    id TEXT PRIMARY KEY,
    range_start INTEGER NOT NULL,
    range_end INTEGER NOT NULL,
    current_pointer INTEGER NOT NULL,
    download_id INTEGER NOT NULL,
    single_part BOOLEAN NOT NULL DEFAULT FALSE,

    FOREIGN KEY (download_id) REFERENCES downloads(id) ON DELETE CASCADE
);
