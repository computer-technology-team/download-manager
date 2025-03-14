CREATE TABLE downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Auto-incrementing primary key
    queue_id INTEGER NOT NULL, -- Foreign key to the Queue table
    FOREIGN KEY (queue_id) REFERENCES queues(id) ON DELETE CASCADE
);
