-- +goose Up
-- +goose StatementBegin
CREATE TABLE notes (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL,
    note_blocks json NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
CREATE INDEX idx_notes_category_id ON notes (category_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notes;
-- +goose StatementEnd
