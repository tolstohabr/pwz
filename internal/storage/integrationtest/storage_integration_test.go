package integrationtest

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"
	"PWZ1.0/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPgStorage_SaveAndGetOrder(t *testing.T) {
	ctx := context.Background()

	db, terminate, err := setupTestDB(ctx)
	require.NoError(t, err)
	defer terminate()

	pgStorage := storage.NewPgStorage(db)

	order := models.Order{
		ID:          1,
		UserID:      10,
		Status:      "ACCEPTED",
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		Weight:      10,
		Price:       1000,
		PackageType: "box",
	}

	err = pgStorage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return pgStorage.SaveOrderTx(ctx, tx, order)
	})
	require.NoError(t, err)

	got, err := pgStorage.GetOrder(ctx, order.ID)
	require.NoError(t, err)

	require.Equal(t, order.ID, got.ID)
	require.Equal(t, order.UserID, got.UserID)
	require.Equal(t, order.Status, got.Status)
	require.WithinDuration(t, order.ExpiresAt.UTC(), got.ExpiresAt.UTC(), time.Second)
	require.Equal(t, order.Weight, got.Weight)
	require.Equal(t, order.Price, got.Price)
	require.Equal(t, order.PackageType, got.PackageType)
}

func TestPgStorage_DeleteOrder(t *testing.T) {
	ctx := context.Background()

	db, terminate, err := setupTestDB(ctx)
	require.NoError(t, err)
	defer terminate()

	pgStorage := storage.NewPgStorage(db)

	order := models.Order{
		ID:          1,
		UserID:      10,
		Status:      "RETURNED",
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		Weight:      10,
		Price:       100,
		PackageType: "box",
	}

	err = pgStorage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return pgStorage.SaveOrderTx(ctx, tx, order)
	})
	require.NoError(t, err)

	err = pgStorage.DeleteOrder(ctx, order.ID)
	require.NoError(t, err)

	_, err = pgStorage.GetOrder(ctx, order.ID)
	require.ErrorIs(t, err, domainErrors.ErrOrderNotFound)
}

func TestPgStorage_UpdateOrderTx(t *testing.T) {
	ctx := context.Background()

	db, terminate, err := setupTestDB(ctx)
	require.NoError(t, err)
	defer terminate()

	pgStorage := storage.NewPgStorage(db)

	order := models.Order{
		ID:          1,
		UserID:      10,
		Status:      "ACCEPTED",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Weight:      10,
		Price:       100,
		PackageType: "box",
	}
	err = pgStorage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return pgStorage.SaveOrderTx(ctx, tx, order)
	})
	require.NoError(t, err)

	updatedOrder := order
	updatedOrder.Status = "RETURNED"
	updatedOrder.Weight = 11
	updatedOrder.Price = 150

	err = pgStorage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return pgStorage.UpdateOrderTx(ctx, tx, updatedOrder)
	})
	require.NoError(t, err)

	got, err := pgStorage.GetOrder(ctx, order.ID)
	require.NoError(t, err)

	require.Equal(t, updatedOrder.Status, got.Status)
	require.Equal(t, updatedOrder.Weight, got.Weight)
	require.Equal(t, updatedOrder.Price, got.Price)

	history, err := pgStorage.GetHistory(ctx, 0, 10)
	require.NoError(t, err)

	found := false
	for _, h := range history {
		if h.OrderID == order.ID && h.Status == "RETURNED" {
			found = true
			break
		}
	}
	require.True(t, found, "ожидалось RETURNED")
}

func TestPgStorage_ListOrders(t *testing.T) {
	ctx := context.Background()

	db, terminate, err := setupTestDB(ctx)
	require.NoError(t, err)
	defer terminate()

	pgStorage := storage.NewPgStorage(db)

	orders := []models.Order{
		{
			ID:          1,
			UserID:      10,
			Status:      "EXPECTS",
			ExpiresAt:   time.Now().Add(24 * time.Hour).UTC(),
			Weight:      10,
			Price:       100,
			PackageType: "box",
		},
		{
			ID:          2,
			UserID:      20,
			Status:      "ACCEPTED",
			ExpiresAt:   time.Now().Add(48 * time.Hour).UTC(),
			Weight:      20,
			Price:       200,
			PackageType: "tape",
		},
	}

	for _, o := range orders {
		err := pgStorage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
			return pgStorage.SaveOrderTx(ctx, tx, o)
		})
		require.NoError(t, err)
	}

	listed, err := pgStorage.ListOrders(ctx)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(listed), len(orders))

	for _, expected := range orders {
		var found bool
		for _, got := range listed {
			if got.ID == expected.ID {
				found = true
				require.Equal(t, expected.UserID, got.UserID)
				require.Equal(t, expected.Status, got.Status)
				require.WithinDuration(t, expected.ExpiresAt.UTC(), got.ExpiresAt.UTC(), time.Second)
				require.Equal(t, expected.Weight, got.Weight)
				require.Equal(t, expected.Price, got.Price)
				require.Equal(t, expected.PackageType, got.PackageType)
				break
			}
		}
		require.True(t, found, "Order ID отсутствуе в списке", expected.ID)
	}
}

func TestPgStorage_GetHistory(t *testing.T) {
	ctx := context.Background()

	db, terminate, err := setupTestDB(ctx)
	require.NoError(t, err)
	defer terminate()

	pgStorage := storage.NewPgStorage(db)

	order := models.Order{
		ID:          1,
		UserID:      10,
		Status:      "EXPECTS",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Weight:      10,
		Price:       100,
		PackageType: "tape",
	}

	err = pgStorage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return pgStorage.SaveOrderTx(ctx, tx, order)
	})
	require.NoError(t, err)

	order.Status = "ACCEPTED"
	err = pgStorage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return pgStorage.UpdateOrderTx(ctx, tx, order)
	})
	require.NoError(t, err)

	history, err := pgStorage.GetHistory(ctx, 0, 10)
	require.NoError(t, err)

	var foundExpects, foundAccepted bool
	for _, h := range history {
		if h.OrderID == order.ID && h.Status == "EXPECTS" {
			foundExpects = true
		}
		if h.OrderID == order.ID && h.Status == "ACCEPTED" {
			foundAccepted = true
		}
	}
	require.True(t, foundExpects, "не хватает статуса expects")
	require.True(t, foundAccepted, "не хватает статуса accepted")
}

func setupTestDB(ctx context.Context) (*pgxpool.Pool, func(), error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_USER":     "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, nil, err
	}
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}

	dsn := "postgres://test:test@" + host + ":" + mappedPort.Port() + "/testdb?sslmode=disable"

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, err
	}

	err = migrateTestDB(pool)
	if err != nil {
		pool.Close()
		container.Terminate(ctx)
		return nil, nil, err
	}

	cleanup := func() {
		pool.Close()
		container.Terminate(ctx)
	}

	return pool, cleanup, nil
}

func migrateTestDB(pool *pgxpool.Pool) error {
	_, filename, _, _ := runtime.Caller(0)
	migrationPath := filepath.Join(filepath.Dir(filename), "test_migrations.sql")

	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return err
	}

	_, err = pool.Exec(context.Background(), string(sqlBytes))
	return err
}
