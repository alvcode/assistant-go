-- +goose Up
-- +goose StatementBegin
CREATE TABLE drive_files(
    id SERIAL PRIMARY KEY,
    drive_struct_id INT NOT NULL REFERENCES drive_structs (id),
    path TEXT NOT NULL,
    ext TEXT,
    size BIGINT NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
CREATE UNIQUE INDEX idx_drive_files_drive_struct_id ON drive_files (drive_struct_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS drive_files;
-- +goose StatementEnd
