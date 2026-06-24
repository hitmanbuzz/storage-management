CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL, -- hashed password
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE files (
    id BIGSERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    extension VARCHAR(20),
    path TEXT NOT NULL,
    size BIGINT NOT NULL CHECK (size >= 0),
    hash BYTEA NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    file_group TEXT,
    file_desc TEXT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_files_user_id ON files(user_id);
CREATE INDEX idx_files_content_hash ON files(hash);
CREATE INDEX idx_files_size ON files(size);
CREATE INDEX idx_files_uploaded_at ON files(uploaded_at DESC);
