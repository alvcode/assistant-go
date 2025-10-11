-- +goose Up
-- +goose StatementBegin
CREATE TABLE drive_file_chunks(
    id SERIAL PRIMARY KEY,
    drive_file_id INT NOT NULL REFERENCES drive_files (id),
    path TEXT NOT NULL,
    size BIGINT NOT NULL,
    chunk_number INTEGER NOT NULL
);
CREATE INDEX idx_drive_file_chunks_file_id ON drive_file_chunks (drive_file_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS drive_file_chunks;
-- +goose StatementEnd
