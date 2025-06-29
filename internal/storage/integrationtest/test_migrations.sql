CREATE TABLE IF NOT EXISTS orders
(
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    status          VARCHAR(20) NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    weight          REAL NOT NULL CHECK (weight > 0),
    total_price     REAL NOT NULL CHECK (total_price >= 0),
    package_type    VARCHAR(20)
    );

CREATE TABLE IF NOT EXISTS order_history
(
    id          BIGSERIAL PRIMARY KEY,
    order_id    BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status      VARCHAR(20) NOT NULL,
    created_at  TIMESTAMP DEFAULT now()
    );