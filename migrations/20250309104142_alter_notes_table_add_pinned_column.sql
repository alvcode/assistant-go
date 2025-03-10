-- +goose Up
-- +goose StatementBegin
ALTER TABLE notes ADD COLUMN pinned BOOLEAN NOT NULL DEFAULT (false);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE notes DROP COLUMN pinned;
-- +goose StatementEnd
