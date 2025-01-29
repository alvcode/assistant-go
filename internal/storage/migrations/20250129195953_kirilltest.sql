-- +goose Up
-- +goose StatementBegin
CREATE TABLE kirill (
   id SERIAL PRIMARY KEY,
   name VARCHAR(100) NOT NULL,
   lastname VARCHAR(200) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS kirill;
-- +goose StatementEnd
