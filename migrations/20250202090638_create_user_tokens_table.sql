-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_tokens (
   user_id INT NOT NULL,
   token VARCHAR(100) NOT NULL,
   refresh_token VARCHAR(100) NOT NULL,
   expired_to INT NOT NULL
);
CREATE UNIQUE INDEX idx_user_tokens_token ON user_tokens (token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_tokens;
-- +goose StatementEnd
