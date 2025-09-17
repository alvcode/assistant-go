-- +goose Up
-- +goose StatementBegin
ALTER TABLE notes ADD COLUMN title VARCHAR(150);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE notes DROP COLUMN title;
-- +goose StatementEnd
