-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_user_tokens_user_id ON user_tokens (user_id);
CREATE INDEX idx_user_tokens_expired_to ON user_tokens (expired_to);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_user_tokens_user_id;
DROP INDEX idx_user_tokens_expired_to;
-- +goose StatementEnd
