-- +goose Up
-- +goose StatementBegin
ALTER TABLE drive_files DROP CONSTRAINT drive_files_drive_struct_id_fkey;
ALTER TABLE drive_files
    ADD CONSTRAINT drive_files_drive_struct_id_fkey
        FOREIGN KEY (drive_struct_id)
            REFERENCES drive_structs(id)
            ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE drive_files DROP CONSTRAINT drive_files_drive_struct_id_fkey;
ALTER TABLE drive_files
    ADD CONSTRAINT drive_files_drive_struct_id_fkey
        FOREIGN KEY (drive_struct_id)
            REFERENCES drive_structs(id);
-- +goose StatementEnd
