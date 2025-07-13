package storage

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	GetOrder(ctx context.Context, id uint64) (models.Order, error)
	DeleteOrder(ctx context.Context, id uint64) error
	ListOrders(ctx context.Context) ([]models.Order, error)
	GetHistory(ctx context.Context, page uint32, count uint32) ([]models.OrderHistory, error)
	SaveOrderTx(ctx context.Context, tx pgx.Tx, order models.Order) error
	UpdateOrderTx(ctx context.Context, tx pgx.Tx, order models.Order) error
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
	//TODO: новая
	SaveEventTx(ctx context.Context, tx pgx.Tx, order models.Event) error
}

type PgStorage struct {
	db *pgxpool.Pool
}

func NewPgStorage(db *pgxpool.Pool) *PgStorage {
	return &PgStorage{db: db}
}

func (ps *PgStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	log.Printf("SQL: %s\nArgs: %v\n", query, args)
}

func (ps *PgStorage) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	start := time.Now()
	log.Println("Starting transaction")

	tx, err := ps.db.Begin(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v\n", err)
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			log.Printf("Transaction panic: %v, rolling back\n", p)
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		log.Printf("Transaction failed: %v, rolling back\n", err)
		_ = tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Failed to commit transaction: %v\n", err)
		return err
	}

	log.Printf("Transaction committed (duration: %v)\n", time.Since(start))
	return nil
}

func (ps *PgStorage) SaveOrderTx(ctx context.Context, tx pgx.Tx, order models.Order) error {
	const query = `
		INSERT INTO orders (id, user_id, status, expires_at, weight, total_price, package_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	ps.logQuery(ctx, query, order.ID, order.UserID, order.Status, order.ExpiresAt, order.Weight, order.Price, order.PackageType)

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
			log.Printf("Duplicate order: %v\n", order.ID)
			return domainErrors.ErrDuplicateOrder
		}
		log.Printf("Failed to save order: %v\n", err)
		return err
	}

	const historyQuery = `
		INSERT INTO order_history (order_id, status)
		VALUES ($1, $2)
	`
	ps.logQuery(ctx, historyQuery, order.ID, order.Status)

	_, err = tx.Exec(ctx, historyQuery, order.ID, order.Status)
	if err != nil {
		log.Printf("Failed to save order history: %v\n", err)
	}
	return err
}

func (ps *PgStorage) SaveEventTx(ctx context.Context, tx pgx.Tx, event models.Event) error {

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v\n", err)
		return err
	}

	const query = `
		INSERT INTO outbox (id, payload, status, created_at)
		VALUES ($1, $2, 'CREATED', now())
	`

	ps.logQuery(ctx, query, event.EventID, string(payload))

	_, err = tx.Exec(ctx, query,
		event.EventID,
		payload,
	)

	if err != nil {
		log.Printf("Failed to insert into outbox: %v\n", err)
		return err
	}

	return nil
}

func (ps *PgStorage) UpdateOrderTx(ctx context.Context, tx pgx.Tx, order models.Order) error {
	const query = `
		UPDATE orders
		SET user_id = $2, status = $3, expires_at = $4, weight = $5,
			total_price = $6, package_type = $7
		WHERE id = $1
	`
	ps.logQuery(ctx, query,
		order.ID,
		order.UserID,
		order.Status,
		order.ExpiresAt,
		order.Weight,
		order.Price,
		order.PackageType,
	)

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
		log.Printf("Failed to update order: %v\n", err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		log.Printf("Order not found for update: %v\n", order.ID)
		return domainErrors.ErrOrderNotFound
	}

	const historyQuery = `
		INSERT INTO order_history (order_id, status)
		VALUES ($1, $2)
	`
	ps.logQuery(ctx, historyQuery, order.ID, order.Status)

	_, err = tx.Exec(ctx, historyQuery, order.ID, order.Status)
	if err != nil {
		log.Printf("Failed to update order history: %v\n", err)
	}
	return err
}

func (ps *PgStorage) GetOrder(ctx context.Context, id uint64) (models.Order, error) {
	const query = `
		SELECT id, user_id, status, expires_at, weight, total_price, package_type
		FROM orders WHERE id = $1
	`
	ps.logQuery(ctx, query, id)

	var order models.Order
	err := ps.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.ExpiresAt,
		&order.Weight,
		&order.Price,
		&order.PackageType,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("Order not found: %v\n", id)
		return models.Order{}, domainErrors.ErrOrderNotFound
	}
	if err != nil {
		log.Printf("Failed to get order: %v\n", err)
	}
	return order, err
}

func (ps *PgStorage) DeleteOrder(ctx context.Context, id uint64) error {
	const query = `DELETE FROM orders WHERE id = $1`
	ps.logQuery(ctx, query, id)

	cmdTag, err := ps.db.Exec(ctx, query, id)
	if err != nil {
		log.Printf("Failed to delete order: %v\n", err)
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		log.Printf("Order not found for deletion: %v\n", id)
		return domainErrors.ErrOrderNotFound
	}

	log.Printf("Order deleted: %v\n", id)
	return nil
}

func (ps *PgStorage) ListOrders(ctx context.Context) ([]models.Order, error) {
	const query = `
		SELECT id, user_id, status, expires_at, weight, total_price, package_type
		FROM orders
	`
	ps.logQuery(ctx, query)

	rows, err := ps.db.Query(ctx, query)
	if err != nil {
		log.Printf("Failed to list orders: %v\n", err)
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
			log.Printf("Failed to scan order row: %v\n", err)
			return nil, err
		}
		orders = append(orders, o)
	}

	log.Printf("Listed %d orders\n", len(orders))
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
	ps.logQuery(ctx, query, count, offset)

	rows, err := ps.db.Query(ctx, query, count, offset)
	if err != nil {
		log.Printf("Failed to get order history: %v\n", err)
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
			log.Printf("Failed to scan history row: %v\n", err)
			return nil, err
		}
		history = append(history, h)
	}

	log.Printf("Retrieved %d history entries\n", len(history))
	return history, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
