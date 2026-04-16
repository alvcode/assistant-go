-- +goose Up
-- +goose StatementBegin
ALTER TABLE drive_recycle_bin ADD COLUMN original_path TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE drive_recycle_bin DROP COLUMN original_path;
-- +goose StatementEnd
