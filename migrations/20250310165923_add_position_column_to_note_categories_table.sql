-- +goose Up
-- +goose StatementBegin
ALTER TABLE note_categories ADD COLUMN position smallint NOT NULL DEFAULT (0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE note_categories DROP COLUMN position;
-- +goose StatementEnd
