-- +goose Up
-- +goose StatementBegin
CREATE TABLE block_events (
    id BIGSERIAL PRIMARY KEY,
    ip INET NOT NULL,
    event VARCHAR(50) NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
CREATE INDEX idx_block_events_ip_created_at ON block_events (ip, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS block_events;
-- +goose StatementEnd
