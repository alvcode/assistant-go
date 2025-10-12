-- +goose Up
-- +goose StatementBegin
ALTER TABLE drive_file_chunks DROP CONSTRAINT drive_file_chunks_drive_file_id_fkey;
ALTER TABLE drive_file_chunks
    ADD CONSTRAINT drive_file_chunks_drive_file_id_fkey
        FOREIGN KEY (drive_file_id)
            REFERENCES drive_files(id)
            ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE drive_file_chunks DROP CONSTRAINT drive_file_chunks_drive_file_id_fkey;
ALTER TABLE drive_file_chunks
    ADD CONSTRAINT drive_file_chunks_drive_file_id_fkey
        FOREIGN KEY (drive_file_id)
            REFERENCES drive_files(id);
-- +goose StatementEnd
