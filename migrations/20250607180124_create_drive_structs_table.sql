-- +goose Up
-- +goose StatementBegin
CREATE TABLE drive_structs(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name TEXT NOT NULL,
    type SMALLINT NOT NULL,
    parent_id INT,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
CREATE INDEX idx_drive_structs_user_id_type ON drive_structs (user_id, type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS drive_structs;
-- +goose StatementEnd
