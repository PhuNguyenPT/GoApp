-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN user_agent TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN user_agent;
-- +goose StatementEnd