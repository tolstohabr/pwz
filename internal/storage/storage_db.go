package storage

import (
	"context"
	"database/sql"
	"errors"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"

	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgStorage struct {
	db *sql.DB
}

func NewPgStorage(db *sql.DB) *PgStorage {
	return &PgStorage{db: db}
}

func (ps *PgStorage) SaveOrder(order models.Order) error {
	const query = `
		INSERT INTO orders (id, user_id, status, expires_at, weight, total_price, package_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
	`
	_, err := ps.db.ExecContext(context.Background(), query,
		order.ID,
		order.UserID,
		order.Status,
		order.ExpiresAt,
		order.Weight,
		order.Price,
		order.PackageType,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domainErrors.ErrDuplicateOrder
		}
		return err
	}

	return nil
}

func (ps *PgStorage) GetOrder(id uint64) (models.Order, error) {
	const query = `
		SELECT id, user_id, status, expires_at, weight, total_price, package_type, created_at, updated_at
		FROM orders WHERE id = $1
	`

	var order models.Order
	err := ps.db.QueryRowContext(context.Background(), query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.ExpiresAt,
		&order.Weight,
		&order.Price,
		&order.PackageType,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Order{}, domainErrors.ErrOrderNotFound
	}
	return order, err
}

func (ps *PgStorage) DeleteOrder(id uint64) error {
	const query = `DELETE FROM orders WHERE id = $1`
	res, err := ps.db.ExecContext(context.Background(), query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domainErrors.ErrOrderNotFound
	}
	return nil
}

func (ps *PgStorage) ListOrders() ([]models.Order, error) {
	const query = `
		SELECT id, user_id, status, expires_at, weight, total_price, package_type, created_at, updated_at
		FROM orders
	`

	rows, err := ps.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.Status,
			&o.ExpiresAt,
			&o.Weight,
			&o.Price,
			&o.PackageType,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func (ps *PgStorage) UpdateOrder(order models.Order) error {
	const query = `
		UPDATE orders
		SET user_id = $2, status = $3, expires_at = $4, weight = $5,
			total_price = $6, package_type = $7, updated_at = now()
		WHERE id = $1
	`
	res, err := ps.db.ExecContext(context.Background(), query,
		order.ID,
		order.UserID,
		order.Status,
		order.ExpiresAt,
		order.Weight,
		order.Price,
		order.PackageType,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domainErrors.ErrOrderNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
