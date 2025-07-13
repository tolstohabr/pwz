-- +goose Up
-- +goose StatementBegin

CREATE TYPE outbox_status AS ENUM ('CREATED', 'PROCESSING', 'COMPLETED', 'FAILED');

CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY,
    payload JSONB NOT NULL,
    status outbox_status NOT NULL,
    error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    sent_at TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS outbox;
DROP TYPE IF EXISTS outbox_status;

-- +goose StatementEnd