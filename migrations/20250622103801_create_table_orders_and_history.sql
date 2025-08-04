-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS orders
(
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    status          SMALLINT NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    weight          REAL NOT NULL CHECK (weight > 0),
    total_price     REAL NOT NULL CHECK (total_price >= 0),
    package_type    SMALLINT,
    created_at      TIMESTAMP DEFAULT now(),
    updated_at      TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_history
(
    id          BIGSERIAL PRIMARY KEY,
    order_id    BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status      SMALLINT NOT NULL,
    created_at  TIMESTAMP DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS order_history;
DROP TABLE IF EXISTS orders;

-- +goose StatementEnd