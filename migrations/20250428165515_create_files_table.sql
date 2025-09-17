-- +goose Up
-- +goose StatementBegin
CREATE TABLE files (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    ext TEXT NOT NULL,
    size INT NOT NULL,
    hash VARCHAR(80) NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
CREATE INDEX idx_files_user_id ON files (user_id);
CREATE UNIQUE INDEX idx_files_hash ON files (hash);
CREATE INDEX idx_files_size ON files (size);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS files;
-- +goose StatementEnd
