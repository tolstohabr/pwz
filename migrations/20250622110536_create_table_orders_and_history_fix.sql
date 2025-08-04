-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders
DROP COLUMN IF EXISTS created_at,
DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders
ADD COLUMN created_at TIMESTAMP DEFAULT NOW(),
ADD COLUMN updated_at TIMESTAMP DEFAULT NOW();
-- +goose StatementEnd