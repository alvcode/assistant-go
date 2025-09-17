-- +goose Up
-- +goose StatementBegin
CREATE TABLE note_categories (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id INT
);
CREATE INDEX idx_note_categories_user_id ON note_categories (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS note_categories;
-- +goose StatementEnd
