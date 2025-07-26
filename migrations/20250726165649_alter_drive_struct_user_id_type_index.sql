-- +goose Up
-- +goose StatementBegin
DROP INDEX idx_drive_structs_user_id_type;
CREATE INDEX idx_drive_structs_user_parent_type ON drive_structs (user_id, parent_id, type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_drive_structs_user_parent_type;
CREATE INDEX idx_drive_structs_user_id_type ON drive_structs (user_id, type);
-- +goose StatementEnd
