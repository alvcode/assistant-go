-- +goose Up
-- +goose StatementBegin
CREATE TABLE drive_recycle_bin (
    id SERIAL PRIMARY KEY,
    drive_struct_id INT NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT drive_rb_drive_struct_id_fkey
        FOREIGN KEY (drive_struct_id)
            REFERENCES drive_structs(id)
            ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_drive_rb_drive_struct_id ON drive_recycle_bin (drive_struct_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_drive_rb_drive_struct_id;
DROP TABLE IF EXISTS drive_recycle_bin;
-- +goose StatementEnd
