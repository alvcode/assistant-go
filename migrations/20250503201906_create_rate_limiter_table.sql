-- +goose Up
-- +goose StatementBegin
CREATE UNLOGGED TABLE rate_limiter(
    ip INET NOT NULL,
    allowance INT NOT NULL,
    timestamp BIGINT NOT NULL
);
CREATE UNIQUE INDEX idx_rate_limiter_ip ON rate_limiter (ip);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rate_limiter;
-- +goose StatementEnd