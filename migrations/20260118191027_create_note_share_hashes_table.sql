-- +goose Up
-- +goose StatementBegin
CREATE TABLE note_share_hashes (
    id SERIAL PRIMARY KEY,
    note_id INT NOT NULL,
    hash VARCHAR(80) NOT NULL,
    CONSTRAINT note_share_hashes_note_id_fkey
        FOREIGN KEY (note_id)
            REFERENCES notes(id)
            ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_note_share_hashes_note_id ON note_share_hashes (note_id);
CREATE UNIQUE INDEX idx_note_share_hashes_hash ON note_share_hashes (hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_note_share_hashes_hash;
DROP INDEX idx_note_share_hashes_note_id;
DROP TABLE IF EXISTS note_share_hashes;
-- +goose StatementEnd
