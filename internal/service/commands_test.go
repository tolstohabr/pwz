package service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"
	"PWZ1.0/internal/storage/mocks"
	"PWZ1.0/internal/tools/logger"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitLogger()
	os.Exit(m.Run())
}

func Test_orderService_AcceptOrder(t *testing.T) {
	type args struct {
		orderID     uint64
		userID      uint64
		weight      float32
		price       float32
		expiresAt   time.Time
		packageType models.PackageType
	}

	tests := []struct {
		name         string
		args         args
		mockSetup    func(m *mocks.StorageMock)
		expectedErr  error
		expectedStat models.OrderStatus
	}{
		{
			name: "successfully creates order",
			args: args{
				orderID:     1,
				userID:      10,
				weight:      10,
				price:       100,
				expiresAt:   time.Now().Add(48 * time.Hour),
				packageType: models.PackageBox,
			},
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					if id == 1 {
						return models.Order{}, domainErrors.ErrOrderNotFound
					}
					return models.Order{}, nil
				})

				m.WithTransactionMock.Set(func(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
					return fn(context.Background(), nil)
				})

				m.SaveOrderTxMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) error {
					if order.ID == 1 &&
						order.UserID == 10 &&
						order.Status == models.StatusExpects {
						return nil
					}
					return errors.New("unexpected order")
				})
			},
			expectedErr:  nil,
			expectedStat: models.StatusExpects,
		},
		{
			name: "expiration date in past",
			args: args{
				orderID:     1,
				userID:      10,
				weight:      10,
				price:       100,
				expiresAt:   time.Now().Add(-2 * time.Hour),
				packageType: models.PackageBox,
			},
			mockSetup: func(m *mocks.StorageMock) {
			},
			expectedErr: domainErrors.ErrValidationFailed,
		},
		{
			name: "order already exists",
			args: args{
				orderID:     1,
				userID:      10,
				weight:      10,
				price:       100,
				expiresAt:   time.Now().Add(48 * time.Hour),
				packageType: models.PackageBox,
			},
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					if id == 1 {
						return models.Order{}, nil
					}
					return models.Order{}, domainErrors.ErrOrderNotFound
				})
			},
			expectedErr: domainErrors.ErrOrderAlreadyExists,
		},
		{
			name: "error saving order",
			args: args{
				orderID:     1,
				userID:      10,
				weight:      10,
				price:       100,
				expiresAt:   time.Now().Add(48 * time.Hour),
				packageType: models.PackageBox,
			},
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					if id == 1 {
						return models.Order{}, domainErrors.ErrOrderNotFound
					}
					return models.Order{}, nil
				})

				m.WithTransactionMock.Set(func(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
					return fn(context.Background(), nil)
				})

				m.SaveOrderTxMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) error {
					return errors.New("db error")
				})
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockStorage := mocks.NewStorageMock(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockStorage)
			}

			svc := NewOrderService(mockStorage)

			order, err := svc.AcceptOrder(
				context.Background(),
				tt.args.orderID,
				tt.args.userID,
				tt.args.weight,
				tt.args.price,
				tt.args.expiresAt,
				tt.args.packageType,
			)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.args.orderID, order.ID)
				assert.Equal(t, tt.expectedStat, order.Status)
			}
		})
	}
}

func Test_orderService_ReturnOrder(t *testing.T) {
	type fields struct {
		storage *mocks.StorageMock
	}
	type args struct {
		ctx     context.Context
		orderID uint64
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		want      *OrderResponse
		wantErr   error
		mockSetup func(m *mocks.StorageMock)
	}{
		{
			name: "order not found",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			want:    nil,
			wantErr: domainErrors.ErrOrderNotFound,
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					return models.Order{}, domainErrors.ErrOrderNotFound
				})
			},
		},
		{
			name: "order already accepted",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			want:    nil,
			wantErr: domainErrors.ErrOrderAlreadyIssued,
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					return models.Order{
						ID:     1,
						Status: models.StatusAccepted,
					}, nil
				})
			},
		},
		{
			name: "order returned and deleted successfully",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			want: &OrderResponse{
				OrderID: 1,
				Status:  models.StatusDeleted,
			},
			wantErr: nil,
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					return models.Order{
						ID:     1,
						Status: models.StatusReturned,
					}, nil
				})
				m.DeleteOrderMock.Set(func(ctx context.Context, id uint64) error {
					return nil
				})
			},
		},
		{
			name: "expired order deleted successfully",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			want: &OrderResponse{
				OrderID: 1,
				Status:  models.StatusDeleted,
			},
			wantErr: nil,
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					return models.Order{
						ID:        1,
						Status:    models.StatusExpects,
						ExpiresAt: time.Now().Add(-time.Hour),
					}, nil
				})
				m.DeleteOrderMock.Set(func(ctx context.Context, id uint64) error {
					return nil
				})
			},
		},
		{
			name: "storage not expired error",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			want:    nil,
			wantErr: domainErrors.ErrStorageNotExpired,
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					return models.Order{
						ID:        1,
						Status:    models.StatusExpects,
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup(tt.fields.storage)
			s := &orderService{
				storage: tt.fields.storage,
			}
			got, err := s.ReturnOrder(tt.args.ctx, tt.args.orderID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_orderService_ProcessOrders(t *testing.T) {
	type fields struct {
		storage *mocks.StorageMock
	}
	type args struct {
		ctx        context.Context
		userID     uint64
		actionType models.ActionType
		orderIDs   []uint64
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		mockSetup func(m *mocks.StorageMock)
		want      ProcessResult
	}{
		{
			name: "successfully process issue action",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:        context.Background(),
				userID:     10,
				actionType: models.ActionTypeIssue,
				orderIDs:   []uint64{1, 2},
			},
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					return models.Order{
						ID:        id,
						UserID:    10,
						Status:    models.StatusExpects,
						ExpiresAt: time.Now().Add(24 * time.Hour),
					}, nil
				})

				m.WithTransactionMock.Set(func(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
					return fn(ctx, nil)
				})

				m.UpdateOrderTxMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) error {
					return nil
				})
			},
			want: ProcessResult{
				Processed: []uint64{1, 2},
				Errors:    []uint64{},
			},
		},
		{
			name: "fail on update error",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:        context.Background(),
				userID:     10,
				actionType: models.ActionTypeIssue,
				orderIDs:   []uint64{1},
			},
			mockSetup: func(m *mocks.StorageMock) {
				m.GetOrderMock.Set(func(ctx context.Context, id uint64) (models.Order, error) {
					return models.Order{
						ID:        id,
						UserID:    10,
						Status:    models.StatusExpects,
						ExpiresAt: time.Now().Add(24 * time.Hour),
					}, nil
				})

				m.WithTransactionMock.Set(func(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
					return fn(ctx, nil)
				})

				m.UpdateOrderTxMock.Set(func(ctx context.Context, tx pgx.Tx, order models.Order) error {
					return errors.New("update failed")
				})
			},
			want: ProcessResult{
				Processed: []uint64{},
				Errors:    []uint64{1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup(tt.fields.storage)
			s := &orderService{
				storage: tt.fields.storage,
			}
			got := s.ProcessOrders(tt.args.ctx, tt.args.userID, tt.args.actionType, tt.args.orderIDs)
			assert.ElementsMatch(t, tt.want.Processed, got.Processed)
			assert.ElementsMatch(t, tt.want.Errors, got.Errors)
		})
	}
}

func Test_orderService_ListOrders(t *testing.T) {
	type fields struct {
		storage *mocks.StorageMock
	}
	type args struct {
		ctx       context.Context
		userID    uint64
		inPvzOnly bool
		lastId    uint32
		page      uint32
		limit     uint32
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		mockSetup  func(m *mocks.StorageMock)
		wantOrders []models.Order
		wantTotal  uint32
		wantErr    bool
	}{
		{
			name: "returns filtered orders with pagination",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:       context.Background(),
				userID:    10,
				inPvzOnly: true,
				lastId:    0,
				page:      0,
				limit:     2,
			},
			mockSetup: func(m *mocks.StorageMock) {
				m.ListOrdersMock.Set(func(ctx context.Context) ([]models.Order, error) {
					return []models.Order{
						{ID: 1, UserID: 10, Status: models.StatusExpects},
						{ID: 2, UserID: 10, Status: models.StatusReturned},
						{ID: 3, UserID: 10, Status: models.StatusAccepted},
						{ID: 4, UserID: 20, Status: models.StatusExpects},
					}, nil
				})
			},
			wantOrders: []models.Order{
				{ID: 1, UserID: 10, Status: models.StatusExpects},
				{ID: 2, UserID: 10, Status: models.StatusReturned},
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name: "limit zero returns empty result and zero total",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:       context.Background(),
				userID:    10,
				inPvzOnly: false,
				lastId:    0,
				page:      0,
				limit:     0,
			},
			mockSetup: func(m *mocks.StorageMock) {
			},
			wantOrders: nil,
			wantTotal:  0,
			wantErr:    false,
		},
		{
			name: "storage list error returns empty and zero",
			fields: fields{
				storage: mocks.NewStorageMock(t),
			},
			args: args{
				ctx:       context.Background(),
				userID:    10,
				inPvzOnly: false,
				lastId:    0,
				page:      0,
				limit:     10,
			},
			mockSetup: func(m *mocks.StorageMock) {
				m.ListOrdersMock.Set(func(ctx context.Context) ([]models.Order, error) {
					return nil, errors.New("storage error")
				})
			},
			wantOrders: []models.Order{},
			wantTotal:  0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup(tt.fields.storage)
			s := &orderService{storage: tt.fields.storage}

			got, got1 := s.ListOrders(tt.args.ctx, tt.args.userID, tt.args.inPvzOnly, tt.args.lastId, tt.args.page, tt.args.limit)

			assert.Equal(t, tt.wantOrders, got)
			assert.Equal(t, tt.wantTotal, got1)
		})
	}
}
