-- +goose Up
-- +goose StatementBegin
CREATE TABLE block_ip (
      id SERIAL PRIMARY KEY,
      ip INET NOT NULL,
      blocked_until TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
CREATE UNIQUE INDEX idx_ip_id ON block_ip (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS block_ip;
-- +goose StatementEnd
