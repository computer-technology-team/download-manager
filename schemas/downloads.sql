CREATE TABLE downloads (
	id BIGSERIAL PRIMARY KEY, 
	url TEXT NOT NULL,
    save_path TEXT NOT NULL,
    bandwidth_limit_bytes_p_s REAL NOT NULL,
    speed_bytes_p_s REAL NOT NULL,
    progress_persent REAL NOT NULL DEFAULT 0.0,
    state TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
