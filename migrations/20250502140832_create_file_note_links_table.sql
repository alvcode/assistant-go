-- +goose Up
-- +goose StatementBegin
CREATE TABLE file_note_links (
    file_id INT NOT NULL,
    note_id INT NOT NULL
);
CREATE INDEX idx_file_note_links_file_id ON file_note_links (file_id);
CREATE INDEX idx_file_note_links_note_id ON file_note_links (note_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS file_note_links;
-- +goose StatementEnd
