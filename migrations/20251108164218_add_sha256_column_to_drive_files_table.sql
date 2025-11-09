-- +goose Up
-- +goose StatementBegin
ALTER TABLE drive_files ADD COLUMN sha256 TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE drive_files DROP COLUMN sha256;
-- +goose StatementEnd
