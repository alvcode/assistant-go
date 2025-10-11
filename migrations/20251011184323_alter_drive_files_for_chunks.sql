-- +goose Up
-- +goose StatementBegin
ALTER TABLE drive_files ALTER COLUMN path DROP NOT NULL;
ALTER TABLE drive_files ADD COLUMN is_chunk BOOLEAN NOT NULL DEFAULT (false);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE drive_files ALTER COLUMN path SET NOT NULL;
ALTER TABLE drive_files DROP COLUMN is_chunk;
-- +goose StatementEnd
