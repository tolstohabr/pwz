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
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PgStorageSuite struct {
	suite.Suite
	db        *pgxpool.Pool
	container testcontainers.Container
	storage   *storage.PgStorage
	ctx       context.Context
	cancel    context.CancelFunc
}

func (s *PgStorageSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Minute)

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
	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(s.T(), err)
	s.container = container

	host, err := container.Host(s.ctx)
	require.NoError(s.T(), err)

	port, err := container.MappedPort(s.ctx, "5432")
	require.NoError(s.T(), err)

	dsn := "postgres://test:test@" + host + ":" + port.Port() + "/testdb?sslmode=disable"
	pool, err := pgxpool.New(s.ctx, dsn)
	require.NoError(s.T(), err)

	err = migrateTestDB(pool)
	require.NoError(s.T(), err)

	s.db = pool
	s.storage = storage.NewPgStorage(pool)
}

func (s *PgStorageSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		_ = s.container.Terminate(s.ctx)
	}
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *PgStorageSuite) TearDownTest() {
	// очищаем все таблицы между тестами
	_, err := s.db.Exec(s.ctx, `
		TRUNCATE TABLE orders CASCADE;
		TRUNCATE TABLE order_history CASCADE;
	`)
	require.NoError(s.T(), err)
}

func (s *PgStorageSuite) Test_SaveAndGetOrder() {
	order := models.Order{
		ID:          1,
		UserID:      10,
		Status:      "ACCEPTED",
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		Weight:      10,
		Price:       1000,
		PackageType: "box",
	}

	err := s.storage.WithTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.storage.SaveOrderTx(ctx, tx, order)
	})
	s.Require().NoError(err)

	got, err := s.storage.GetOrder(s.ctx, order.ID)
	s.Require().NoError(err)

	s.Require().Equal(order.ID, got.ID)
	s.Require().Equal(order.UserID, got.UserID)
	s.Require().Equal(order.Status, got.Status)
	s.Require().WithinDuration(order.ExpiresAt.UTC(), got.ExpiresAt.UTC(), time.Second)
	s.Require().Equal(order.Weight, got.Weight)
	s.Require().Equal(order.Price, got.Price)
	s.Require().Equal(order.PackageType, got.PackageType)
}

func (s *PgStorageSuite) Test_DeleteOrder() {
	order := models.Order{
		ID:          1,
		UserID:      10,
		Status:      "RETURNED",
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		Weight:      10,
		Price:       100,
		PackageType: "box",
	}

	err := s.storage.WithTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.storage.SaveOrderTx(ctx, tx, order)
	})
	s.Require().NoError(err)

	err = s.storage.DeleteOrder(s.ctx, order.ID)
	s.Require().NoError(err)

	_, err = s.storage.GetOrder(s.ctx, order.ID)
	s.Require().ErrorIs(err, domainErrors.ErrOrderNotFound)
}

func (s *PgStorageSuite) Test_UpdateOrderTx() {
	order := models.Order{
		ID:          1,
		UserID:      10,
		Status:      "ACCEPTED",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Weight:      10,
		Price:       100,
		PackageType: "box",
	}
	err := s.storage.WithTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.storage.SaveOrderTx(ctx, tx, order)
	})
	s.Require().NoError(err)

	updatedOrder := order
	updatedOrder.Status = "RETURNED"
	updatedOrder.Weight = 11
	updatedOrder.Price = 150

	err = s.storage.WithTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.storage.UpdateOrderTx(ctx, tx, updatedOrder)
	})
	s.Require().NoError(err)

	got, err := s.storage.GetOrder(s.ctx, order.ID)
	s.Require().NoError(err)

	s.Require().Equal(updatedOrder.Status, got.Status)
	s.Require().Equal(updatedOrder.Weight, got.Weight)
	s.Require().Equal(updatedOrder.Price, got.Price)

	history, err := s.storage.GetHistory(s.ctx, 0, 10)
	s.Require().NoError(err)

	found := false
	for _, h := range history {
		if h.OrderID == order.ID && h.Status == "RETURNED" {
			found = true
			break
		}
	}
	s.Require().True(found, "ожидалось RETURNED в истории")
}

func (s *PgStorageSuite) Test_ListOrders() {
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
		err := s.storage.WithTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
			return s.storage.SaveOrderTx(ctx, tx, o)
		})
		s.Require().NoError(err)
	}

	listed, err := s.storage.ListOrders(s.ctx)
	s.Require().NoError(err)

	s.Require().GreaterOrEqual(len(listed), len(orders))

	for _, expected := range orders {
		var found bool
		for _, got := range listed {
			if got.ID == expected.ID {
				found = true
				s.Require().Equal(expected.UserID, got.UserID)
				s.Require().Equal(expected.Status, got.Status)
				s.Require().WithinDuration(expected.ExpiresAt.UTC(), got.ExpiresAt.UTC(), time.Second)
				s.Require().Equal(expected.Weight, got.Weight)
				s.Require().Equal(expected.Price, got.Price)
				s.Require().Equal(expected.PackageType, got.PackageType)
				break
			}
		}
		s.Require().True(found, "Order ID отсутствует в списке", expected.ID)
	}
}

func (s *PgStorageSuite) Test_GetHistory() {
	order := models.Order{
		ID:          1,
		UserID:      10,
		Status:      "EXPECTS",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Weight:      10,
		Price:       100,
		PackageType: "tape",
	}

	err := s.storage.WithTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.storage.SaveOrderTx(ctx, tx, order)
	})
	s.Require().NoError(err)

	order.Status = "ACCEPTED"
	err = s.storage.WithTransaction(s.ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.storage.UpdateOrderTx(ctx, tx, order)
	})
	s.Require().NoError(err)

	history, err := s.storage.GetHistory(s.ctx, 0, 10)
	s.Require().NoError(err)

	var foundExpects, foundAccepted bool
	for _, h := range history {
		if h.OrderID == order.ID && h.Status == "EXPECTS" {
			foundExpects = true
		}
		if h.OrderID == order.ID && h.Status == "ACCEPTED" {
			foundAccepted = true
		}
	}
	s.Require().True(foundExpects, "не хватает статуса expects")
	s.Require().True(foundAccepted, "не хватает статуса accepted")
}

func TestPgStorageSuite(t *testing.T) {
	suite.Run(t, new(PgStorageSuite))
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
