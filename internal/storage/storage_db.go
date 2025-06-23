package storage

import (
	"context"
	"errors"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	GetOrder(id uint64) (models.Order, error)
	DeleteOrder(id uint64) error
	ListOrders() ([]models.Order, error)
	GetHistory(ctx context.Context, page uint32, count uint32) ([]models.OrderHistory, error)
	SaveOrderTx(ctx context.Context, tx pgx.Tx, order models.Order) error
	UpdateOrderTx(ctx context.Context, tx pgx.Tx, order models.Order) error
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

func (ps *PgStorage) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := ps.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (ps *PgStorage) SaveOrderTx(ctx context.Context, tx pgx.Tx, order models.Order) error {
	const query = `
		INSERT INTO orders (id, user_id, status, expires_at, weight, total_price, package_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(ctx, query,
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

	const historyQuery = `
		INSERT INTO order_history (order_id, status)
		VALUES ($1, $2)
	`
	_, err = tx.Exec(ctx, historyQuery, order.ID, order.Status)
	return err
}

func (ps *PgStorage) UpdateOrderTx(ctx context.Context, tx pgx.Tx, order models.Order) error {
	const query = `
		UPDATE orders
		SET user_id = $2, status = $3, expires_at = $4, weight = $5,
			total_price = $6, package_type = $7
		WHERE id = $1
	`
	cmdTag, err := tx.Exec(ctx, query,
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
	if cmdTag.RowsAffected() == 0 {
		return domainErrors.ErrOrderNotFound
	}

	const historyQuery = `
		INSERT INTO order_history (order_id, status)
		VALUES ($1, $2)
	`
	_, err = tx.Exec(ctx, historyQuery, order.ID, order.Status)
	return err
}

type PgStorage struct {
	db *pgxpool.Pool
}

func NewPgStorage(db *pgxpool.Pool) *PgStorage {
	return &PgStorage{db: db}
}

func (ps *PgStorage) GetOrder(id uint64) (models.Order, error) {
	const query = `
		SELECT id, user_id, status, expires_at, weight, total_price, package_type
		FROM orders WHERE id = $1
	`

	var order models.Order
	err := ps.db.QueryRow(context.Background(), query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.ExpiresAt,
		&order.Weight,
		&order.Price,
		&order.PackageType,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Order{}, domainErrors.ErrOrderNotFound
	}
	return order, err
}

func (ps *PgStorage) DeleteOrder(id uint64) error {
	const query = `DELETE FROM orders WHERE id = $1`
	cmdTag, err := ps.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domainErrors.ErrOrderNotFound
	}
	return nil
}

func (ps *PgStorage) ListOrders() ([]models.Order, error) {
	const query = `
		SELECT id, user_id, status, expires_at, weight, total_price, package_type
		FROM orders
	`

	rows, err := ps.db.Query(context.Background(), query)
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

	return orders, rows.Err()
}

func (ps *PgStorage) GetHistory(ctx context.Context, page, count uint32) ([]models.OrderHistory, error) {
	if count == 0 {
		count = 50
	}
	offset := page * count

	const query = `
		SELECT id, order_id, status, created_at
		FROM order_history
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := ps.db.Query(ctx, query, count, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.OrderHistory
	for rows.Next() {
		var h models.OrderHistory
		err := rows.Scan(
			&h.ID,
			&h.OrderID,
			&h.Status,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
