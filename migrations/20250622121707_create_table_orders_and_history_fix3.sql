-- +goose Up
-- +goose StatementBegin

ALTER TABLE orders
ALTER COLUMN package_type TYPE VARCHAR(20) USING
    CASE package_type
      WHEN 0 THEN 'unspecified'
      WHEN 1 THEN 'bag'
      WHEN 2 THEN 'box'
      WHEN 3 THEN 'tape'
      WHEN 4 THEN 'bag+tape'
      WHEN 5 THEN 'box+tape'
      ELSE 'unspecified'
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE orders
ALTER COLUMN package_type TYPE SMALLINT USING
    CASE package_type
      WHEN 'unspecified' THEN 0
      WHEN 'bag' THEN 1
      WHEN 'box' THEN 2
      WHEN 'tape' THEN 3
      WHEN 'bag+tape' THEN 4
      WHEN 'box+tape' THEN 5
      ELSE 0
END;

-- +goose StatementEnd