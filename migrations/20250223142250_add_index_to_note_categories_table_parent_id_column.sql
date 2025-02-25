-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_note_categories_parent_id ON note_categories (parent_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_note_categories_parent_id;
-- +goose StatementEnd
